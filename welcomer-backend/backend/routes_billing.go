package backend

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"hash/crc32"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
	"github.com/plutov/paypal/v4"
)

// ISO 3166-1 alpha-2 country codes for the Eurozone.
var euroZone = []string{"AT", "BE", "HR", "CY", "EE", "FI", "FR", "DE", "GR", "IE", "IT", "LV", "LT", "LU", "MT", "NL", "PT", "SK", "SI", "ES"}

var certificateCache map[string][]byte = make(map[string][]byte)

func getAvailableCurrencies(ipintelResponse utils.IPIntelResponse) []welcomer.Currency {
	// If the IPIntel response is above the threshold, we assume the user is on a VPN.
	if ipintelResponse.Result <= IPIntelThreshold {
		mapping, ok := welcomer.CountryMapping[ipintelResponse.Country]
		if ok {
			return append(welcomer.GlobalCurrencies, mapping)
		}
	}

	return welcomer.GlobalCurrencies
}

func getDefaultCurrency(ipIntelResponse utils.IPIntelResponse) welcomer.Currency {
	switch {
	case ipIntelResponse.Country == "GB":
		return welcomer.CurrencyGBP
	case ipIntelResponse.Country == "IN":
		return welcomer.CurrencyINR
	case utils.SliceContains(euroZone, ipIntelResponse.Country):
		return welcomer.CurrencyEUR
	default:
		return welcomer.CurrencyUSD
	}
}

type GetSKUsResponse struct {
	AvailableCurrencies []welcomer.Currency   `json:"available_currencies"`
	DefaultCurrency     welcomer.Currency     `json:"default_currency"`
	SKUs                []welcomer.PricingSKU `json:"skus"`
}

func getSKUPricing() map[welcomer.SKUName]welcomer.PricingSKU {
	index, err := strconv.Atoi(os.Getenv("PRICING_TABLE"))
	if err != nil || index < 0 || index >= len(welcomer.SKUPricingTable) {
		backend.Logger.Error().Err(err).Msg("Invalid PRICING_TABLE environment variable")

		return nil
	}

	return welcomer.SKUPricingTable[index]
}

// Route GET /api/billing/skus.
func getSKUs(ctx *gin.Context) {
	response, err := backend.IPChecker.CheckIP(ctx, ctx.ClientIP(), utils.IPIntelFlagDynamicBanListDynamicChecks, utils.IPIntelOFlagShowCountry)
	if err != nil {
		backend.Logger.Warn().Err(err).IPAddr("ip", net.IP(ctx.ClientIP())).Msg("Failed to validate IP via IPIntel")
	}

	currencies := getAvailableCurrencies(response)

	skuPricing := getSKUPricing()

	pricingStructure := GetSKUsResponse{
		AvailableCurrencies: currencies,
		DefaultCurrency:     getDefaultCurrency(response),
		SKUs:                make([]welcomer.PricingSKU, 0, len(skuPricing)),
	}

	for _, sku := range skuPricing {
		pricingStructure.SKUs = append(pricingStructure.SKUs, sku)
	}

	ctx.JSON(http.StatusOK, BaseResponse{
		Ok:   true,
		Data: pricingStructure,
	})
}

type CreatePaymentRequest struct {
	SKU      welcomer.SKUName  `json:"sku"`
	Currency welcomer.Currency `json:"currency"`
}

type CreatePaymentResponse struct {
	URL string `json:"url"`
}

// Route POST /api/billing/payments.
func createPayment(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		// Read "request" text from the post json body.
		var request CreatePaymentRequest

		if err := ctx.ShouldBindJSON(&request); err != nil {
			backend.Logger.Warn().Err(err).Msg("Failed to bind JSON")

			ctx.JSON(http.StatusBadRequest, NewBaseResponse(ErrInvalidJSON, nil))

			return
		}

		if request.SKU == "" {
			ctx.JSON(http.StatusBadRequest, NewBaseResponse(NewMissingParameterError("sku"), nil))

			return
		}

		skus := getSKUPricing()

		sku, ok := skus[request.SKU]
		if !ok {
			backend.Logger.Warn().Str("sku", string(request.SKU)).Msg("Invalid SKU")

			ctx.JSON(http.StatusBadRequest, NewBaseResponse(ErrInvalidSKU, nil))

			return
		}

		response, err := backend.IPChecker.CheckIP(ctx, ctx.ClientIP(), utils.IPIntelFlagDynamicBanListDynamicChecks, utils.IPIntelOFlagShowCountry)
		if err != nil {
			backend.Logger.Warn().Err(err).IPAddr("ip", net.IP(ctx.ClientIP())).Msg("Failed to validate IP via IPIntel")
		}

		if request.Currency == "" {
			request.Currency = getDefaultCurrency(response)
		}

		currencies := getAvailableCurrencies(response)
		if !utils.SliceContains(currencies, request.Currency) {
			backend.Logger.Warn().Str("currency", string(request.Currency)).Msg("Invalid currency")

			ctx.JSON(http.StatusBadRequest, NewBaseResponse(ErrCurrencyNotAvailable, nil))

			return
		}

		// Check if the currency is available for this SKU.
		skuCost, ok := sku.Costs[request.Currency]
		if !ok {
			backend.Logger.Warn().Str("currency", string(request.Currency)).Msg("Invalid currency")

			ctx.JSON(http.StatusBadRequest, NewBaseResponse(ErrCurrencyNotAvailableForSKU, nil))

			return
		}

		// Check if the cost is valid.
		if utils.TryParseFloat(skuCost) <= 0 {
			backend.Logger.Warn().Str("currency", string(request.Currency)).Str("sku", string(request.SKU)).Msg("Invalid cost")

			ctx.JSON(http.StatusInternalServerError, NewBaseResponse(ErrInvalidCost, nil))

			return
		}

		user := tryGetUser(ctx)

		// Create a transaction for the user.
		money := &paypal.Money{
			Currency: string(request.Currency),
			Value:    skuCost,
		}

		applicationContext := &paypal.ApplicationContext{
			BrandName:          "Welcomer",
			ShippingPreference: paypal.ShippingPreferenceNoShipping,
			UserAction:         paypal.UserActionPayNow,
			PaymentMethod: paypal.PaymentMethod{
				PayeePreferred:         paypal.PayeePreferredUnrestricted,
				StandardEntryClassCode: paypal.StandardEntryClassCodeWeb,
			},
			LandingPage: "NO_PREFERENCE",
			ReturnURL:   "https://" + backend.Options.Domain + "/api/billing/callback",
			CancelURL:   "https://" + backend.Options.Domain + "/premium#cancelled",
		}

		if sku.IsRecurring {
			subscriptionID, ok := sku.PaypalSubscriptionID[request.Currency]
			if ok && subscriptionID != "" {
				applicationContext.UserAction = paypal.UserActionSubscribeNow
				createPaymentSubscription(ctx, sku, applicationContext, user, money, subscriptionID)

				return
			}
		}

		createPaymentOrder(ctx, sku, applicationContext, user, money)
	})
}

func createPaymentSubscription(ctx *gin.Context, sku welcomer.PricingSKU, applicationContext *paypal.ApplicationContext, user SessionUser, money *paypal.Money, paypalSubscriptionID string) {
	// Create a user transaction.
	userTransaction, err := welcomer.CreateTransactionForUser(
		ctx,
		backend.Database,
		user.ID,
		database.PlatformTypePaypalSubscription,
		database.TransactionStatusPending,
		"",
		money.Currency,
		money.Value,
	)
	if err != nil {
		backend.Logger.Error().Err(err).Msg("Failed to create user transaction")

		ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

		return
	}

	subscriptionBase := paypal.SubscriptionBase{
		PlanID:   paypalSubscriptionID,
		Quantity: "1",
		Subscriber: &paypal.Subscriber{
			PayerID: user.ID.String(),
			Name: paypal.CreateOrderPayerName{
				GivenName: user.GlobalName,
			},
			ShippingAddress: paypal.ShippingDetail{
				Name: &paypal.Name{
					FullName: user.GlobalName,
				},
				Address: &paypal.ShippingDetailAddressPortable{
					AddressLine1: "Unknown",
					CountryCode:  "US",
				},
			},
		},
		AutoRenewal:        true,
		ApplicationContext: applicationContext,
		CustomID:           userTransaction.TransactionUuid.String(),
		Plan: &paypal.PlanOverride{
			BillingCycles: []paypal.BillingCycleOverride{
				{
					Sequence:    utils.ToPointer(1),
					TotalCycles: utils.ToPointer(1),
					PricingScheme: paypal.PricingScheme{
						FixedPrice: paypal.Money{
							Currency: money.Currency,
							Value:    "0",
						},
					},
				},
				{
					Sequence:    utils.ToPointer(2),
					TotalCycles: utils.ToPointer(0),
					PricingScheme: paypal.PricingScheme{
						FixedPrice: *money,
					},
				},
			},
		},
	}

	// Send subscription request to paypal.
	subscription, err := backend.PaypalClient.CreateSubscription(ctx, subscriptionBase)
	if err != nil {
		backend.Logger.Error().Err(err).Msg("Failed to create subscription")

		ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

		return
	}

	backend.Logger.Info().
		Str("subscriptionID", subscription.ID).
		Str("transactionUuid", userTransaction.TransactionUuid.String()).
		Msg("Created subscription")

	// Update the user transaction with the subscription ID.
	userTransaction.TransactionID = subscription.ID

	_, err = backend.Database.UpdateUserTransaction(ctx, database.UpdateUserTransactionParams{
		TransactionUuid:   userTransaction.TransactionUuid,
		UserID:            userTransaction.UserID,
		PlatformType:      userTransaction.PlatformType,
		TransactionID:     userTransaction.TransactionID,
		TransactionStatus: userTransaction.TransactionStatus,
		CurrencyCode:      userTransaction.CurrencyCode,
		Amount:            userTransaction.Amount,
	})
	if err != nil {
		backend.Logger.Error().Err(err).Msg("Failed to update user transaction")

		ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

		return
	}

	// Get the payer action link.
	payerActionLink := getLinkNamed(subscription.Links, "approve")
	if payerActionLink == "" {
		backend.Logger.Error().Msg("Failed to get payer action link from subscription response")

		ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

		return
	}

	ctx.JSON(http.StatusOK, BaseResponse{
		Ok: true,
		Data: CreatePaymentResponse{
			URL: payerActionLink,
		},
	})
}

func createPaymentOrder(ctx *gin.Context, sku welcomer.PricingSKU, applicationContext *paypal.ApplicationContext, user SessionUser, money *paypal.Money) {

	// Create a user transaction.
	userTransaction, err := welcomer.CreateTransactionForUser(
		ctx,
		backend.Database,
		user.ID,
		database.PlatformTypePaypal,
		database.TransactionStatusPending,
		"",
		money.Currency,
		money.Value,
	)
	if err != nil {
		backend.Logger.Error().Err(err).Msg("Failed to create user transaction")

		ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

		return
	}

	// Create a purchase unit for the user.
	purchaseUnit := paypal.PurchaseUnitRequest{
		Amount: &paypal.PurchaseUnitAmount{
			Currency: money.Currency,
			Value:    money.Value,
			Breakdown: &paypal.PurchaseUnitAmountBreakdown{
				ItemTotal: money,
			},
		},
		Payee:          nil,
		Description:    sku.Description,
		CustomID:       userTransaction.TransactionUuid.String(),
		SoftDescriptor: sku.SoftDescriptor,
		Items: []paypal.Item{
			{
				Name:        sku.Name,
				UnitAmount:  money,
				Tax:         nil,
				Quantity:    "1",
				Description: sku.Description,
				SKU:         string(sku.ID),
				Category:    paypal.ItemCategoryDigitalGood,
			},
		},
		Shipping:           nil,
		PaymentInstruction: nil,
	}

	// Send order request to paypal.
	order, err := backend.PaypalClient.CreateOrder(ctx, paypal.OrderIntentCapture, []paypal.PurchaseUnitRequest{purchaseUnit}, nil, applicationContext)
	if err != nil {
		backend.Logger.Error().Err(err).Msg("Failed to create order")

		ctx.JSON(http.StatusInternalServerError, NewBaseResponse(ErrCreateOrderFailed, nil))

		return
	}

	backend.Logger.Info().
		Str("orderID", order.ID).
		Str("transactionUuid", userTransaction.TransactionUuid.String()).
		Msg("Created order")

	// Update the user transaction with the order ID.
	userTransaction.TransactionID = order.ID

	_, err = backend.Database.UpdateUserTransaction(ctx, database.UpdateUserTransactionParams{
		TransactionUuid:   userTransaction.TransactionUuid,
		UserID:            userTransaction.UserID,
		PlatformType:      userTransaction.PlatformType,
		TransactionID:     userTransaction.TransactionID,
		TransactionStatus: userTransaction.TransactionStatus,
		CurrencyCode:      userTransaction.CurrencyCode,
		Amount:            userTransaction.Amount,
	})
	if err != nil {
		backend.Logger.Error().Err(err).Msg("Failed to update user transaction")

		ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

		return
	}

	// Get the payer action link.
	payerActionLink := getLinkNamed(order.Links, "approve")
	if payerActionLink == "" {
		backend.Logger.Error().Msg("Failed to get payer action link from order response")

		ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

		return
	}

	ctx.JSON(http.StatusOK, BaseResponse{
		Ok: true,
		Data: CreatePaymentResponse{
			URL: payerActionLink,
		},
	})
}

func getLinkNamed(links []paypal.Link, name string) string {
	for _, link := range links {
		if link.Rel == name {
			return link.Href
		}
	}

	return ""
}

// Route POST /api/billing/cancelled?token=...
func paymentCancelled(ctx *gin.Context) {
	ctx.Header("Location", "https://"+backend.Options.Domain+"/premium#cancelled")
	ctx.Status(http.StatusTemporaryRedirect)
}

// Route POST /api/billing/callback?token=...&PayerID=...
func paymentCallback(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		err := backend.Pool.BeginTxFunc(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
			queries := backend.Database.WithTx(tx)

			// Read "token" and "PayerID" from the query string.
			token := ctx.Query("token")
			payerID := ctx.Query("PayerID")

			if token == "" {
				backend.Logger.Warn().Str("PayerID", payerID).Msg("Missing token")

				ctx.JSON(http.StatusBadRequest, NewBaseResponse(NewMissingParameterError("token"), nil))

				return ErrMissingParameter
			}

			if payerID == "" {
				backend.Logger.Warn().Str("token", token).Msg("Missing PayerID")

				ctx.JSON(http.StatusBadRequest, NewBaseResponse(NewMissingParameterError("PayerID"), nil))

				return ErrMissingParameter
			}

			transactions, err := queries.GetUserTransactionsByTransactionID(ctx, token)
			if err != nil {
				backend.Logger.Error().Err(err).Str("token", token).Msg("Failed to get user transactions by transaction ID")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

				return err
			}

			if len(transactions) == 0 {
				backend.Logger.Warn().Str("token", token).Msg("No user transactions found")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

				return err
			}

			user := tryGetUser(ctx)
			transaction := transactions[0]

			if transaction.UserID != int64(user.ID) {
				backend.Logger.Warn().Str("token", token).Int64("userID", transaction.UserID).Int64("user.ID", int64(user.ID)).Msg("User ID does not match")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

				return err
			}

			if transaction.TransactionStatus != int32(database.TransactionStatusPending) {
				backend.Logger.Warn().Str("token", token).Str("transactionStatus", database.TransactionStatus(transaction.TransactionStatus).String()).Msg("Transaction is not pending")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

				return err
			}

			// Get order
			order, err := backend.PaypalClient.GetOrder(ctx, token)
			if err != nil {
				backend.Logger.Error().Err(err).Str("token", token).Msg("Failed to get order")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(ErrGetOrderValidationFailed, nil))

				return err
			}

			if len(order.PurchaseUnits) == 0 || len(order.PurchaseUnits[0].Items) == 0 {
				backend.Logger.Warn().Str("token", token).Msg("No purchase units or items found")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(ErrGetOrderValidationFailed, nil))

				return err
			}

			// Fetch SKU from the order.
			skuName := order.PurchaseUnits[0].Items[0].SKU

			pricing := getSKUPricing()

			sku, ok := pricing[welcomer.SKUName(skuName)]
			if !ok {
				backend.Logger.Warn().Str("sku", skuName).Msg("Invalid SKU")

				ctx.JSON(http.StatusBadRequest, NewBaseResponse(ErrInvalidSKU, nil))

				return err
			}

			// Capture the order
			authorizeResponse, err := backend.PaypalClient.CaptureOrder(ctx, token, paypal.CaptureOrderRequest{})
			if err != nil || authorizeResponse.Status != paypal.OrderStatusCompleted {
				backend.Logger.Error().Err(err).Str("token", token).Str("status", authorizeResponse.Status).Msg("Failed to authorize order")

				// Create a user transaction.
				_, err = welcomer.CreateTransactionForUser(
					ctx,
					queries,
					user.ID,
					database.PlatformTypePaypal,
					database.TransactionStatusPending,
					authorizeResponse.ID,
					transaction.CurrencyCode,
					transaction.Amount,
				)
				if err != nil {
					backend.Logger.Error().Err(err).Msg("Failed to create user transaction")
				}

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(ErrCaptureOrderFailed, nil))

				return err
			}

			// Create a user transaction.
			userTransaction, err := welcomer.CreateTransactionForUser(
				ctx,
				queries,
				user.ID,
				database.PlatformTypePaypal,
				database.TransactionStatusCompleted,
				authorizeResponse.ID,
				transaction.CurrencyCode,
				transaction.Amount,
			)
			if err != nil {
				backend.Logger.Error().Err(err).Msg("Failed to create user transaction")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(ErrCreateTransactionFailed, nil))

				return err
			}

			startedAt := time.Time{}
			expiresAt := startedAt.AddDate(0, utils.If(sku.MonthCount < 0, 120, sku.MonthCount), 0)

			// Create a new membership for the user.
			err = welcomer.CreateMembershipForUser(
				ctx,
				queries,
				user.ID,
				userTransaction.TransactionUuid,
				sku.MembershipType,
				expiresAt,
				nil,
			)
			if err != nil {
				backend.Logger.Error().Err(err).Msg("Failed to create new membership")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(ErrCreateMembershipFailed, nil))

				return err
			}

			ctx.Header("Location", "https://"+backend.Options.Domain+"/premium#success")
			ctx.Status(http.StatusTemporaryRedirect)

			return nil
		})
		if err != nil && !ctx.Writer.Written() {
			backend.Logger.Error().Err(err).Msg("Failed to process payment")

			ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))
		}
	})
}

func downloadAndCache(url string) ([]byte, error) {
	if body, ok := certificateCache[url]; ok {
		return body, nil
	}

	// Download the certificate.
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	certificateCache[url] = body

	return body, nil
}

func validatePaypalWebhook(ctx *gin.Context) ([]byte, bool) {
	headers := ctx.Request.Header

	transmissionID := headers.Get("Paypal-Transmission-Id")
	timeStamp := headers.Get("Paypal-Transmission-Time")
	certURL := headers.Get("Paypal-Cert-Url")

	if transmissionID == "" || timeStamp == "" || certURL == "" {
		backend.Logger.Error().Msg("Missing headers")

		return nil, false
	}

	event, err := ctx.GetRawData()
	if err != nil {
		backend.Logger.Error().Err(err).Msg("Failed to read event data")

		return nil, false
	}
	crc := crc32.ChecksumIEEE(event)
	message := fmt.Sprintf("%s|%s|%s|%d", transmissionID, timeStamp, "WEBHOOK_ID", crc)

	certPemBytes, err := downloadAndCache(certURL)
	if err != nil {
		backend.Logger.Error().Err(err).Msg("Failed to download and cache certificate")

		return nil, false
	}

	block, _ := pem.Decode(certPemBytes)
	if block == nil {
		backend.Logger.Error().Msg("Failed to decode PEM block containing the certificate")

		return nil, false
	}

	certPem, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		backend.Logger.Error().Err(err).Msg("Failed to decode certificate")

		return nil, false
	}

	signature, err := base64.StdEncoding.DecodeString(headers.Get("Paypal-Transmission-Sig"))
	if err != nil {
		backend.Logger.Error().Err(err).Msg("Failed to decode signature")

		return nil, false
	}

	hashed := crypto.SHA256.New()
	hashed.Write([]byte(message))

	pub, ok := certPem.PublicKey.(*rsa.PublicKey)
	if !ok {
		backend.Logger.Error().Msg("Failed to get public key")

		return nil, false
	}

	err = rsa.VerifyPKCS1v15(pub, crypto.SHA256, hashed.Sum(nil), signature)
	if err != nil {
		backend.Logger.Error().Msg("Failed to verify signature")

		return nil, false
	}

	return event, true
}

// Route POST /api/billing/paypal_webhook
func paypalWebhook(ctx *gin.Context) {
	event, ok := validatePaypalWebhook(ctx)
	if !ok {
		ctx.JSON(http.StatusForbidden, NewBaseResponse(ErrWebhookValidationFailed, nil))

		return
	}

	// Handle different event types
	var webhookEvent paypal.Event
	if err := json.Unmarshal(event, &webhookEvent); err != nil {
		backend.Logger.Error().Err(err).Msg("Failed to unmarshal webhook event")
		ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))
		return
	}

	switch webhookEvent.EventType {
	case "BILLING.SUBSCRIPTION.CREATED":
		// Handle subscription created
	case "BILLING.SUBSCRIPTION.ACTIVATED":
		// Handle subscription activated
	case "BILLING.SUBSCRIPTION.EXPIRED":
		// Handle subscription expired
	case "BILLING.SUBSCRIPTION.CANCELLED":
		// Handle subscription cancelled
	case "BILLING.SUBSCRIPTION.SUSPENDED":
		// Handle subscription suspended
	case "BILLING.SUBSCRIPTION.PAYMENT.FAILED":
		// Handle subscription payment failed
	default:
		backend.Logger.Warn().Str("event_type", webhookEvent.EventType).Msg("Unhandled event type")
	}

	ctx.JSON(http.StatusOK, BaseResponse{Ok: true})

	// BILLING.SUBSCRIPTION.CREATED
	// BILLING.SUBSCRIPTION.ACTIVATED
	// BILLING.SUBSCRIPTION.EXPIRED
	// BILLING.SUBSCRIPTION.CANCELLED
	// BILLING.SUBSCRIPTION.SUSPENDED
	// BILLING.SUBSCRIPTION.PAYMENT.FAILED

	// https://beta.welcomer.gg/api/billing/callback?subscription_id=I-TMR758P4PW2C&ba_token=BA-1UW93399KE424124U&token=6YC07506GL255092S

	// Headers
	// Cf-Timezone America/Chicago
	// Paypal-Auth-Version v2
	// Cf-Iplatitude 37.75100
	// Paypal-Transmission-Id a05ce04f-f7bf-11ef-be09-87c28fdcff97
	// Paypal-Cert-Url https://api.paypal.com/v1/notifications/certs/CERT-360caa42-fca2a594-b0d12406
	// X-Real-Ip 172.69.34.129
	// Cf-Ray 91a4bd36c8a152b3-LAX
	// X-B3-Spanid c012cba11785d09b
	// Paypal-Transmission-Sig fzvDj44jMCaBlyMkSJQbC+RkGjGlE492u9PAVaqDvNHm3BHez5PNHIVCblKB4+tdkjADWPhgkqUfn3RRy9TSZLNpsooqgEWDxWQ+ccW0LJweXQt7B07ihXgjXDEeElqApGAih9EBgjX1TIbRGDitUau4d9uCITGIwwkm1xC/rvqT2UW1dvq11/TEJVEUR3FQmzveiumNP1sAEb9CmpqKYUA2GW6tfzIHIoBxdvTh3rCm1Ehw2zypk4Zm537hpc3i9gy1qPc/Ik8V5tlPMfPOwwC9/PIQgq5GKUInGtcMclYXFOeYyigU2SKGiOBQ2Qc++j5PCm3SiE3xktrYDsLWrw==
	// Correlation-Id e3bb88e4c2d53
	// Content-Length 2305
	// Cdn-Loop cloudflare; loops=1
	// Paypal-Auth-Algo SHA256withRSA
	// Cf-Ipcontinent NA
	// Cf-Iplongitude -97.82200
	// Accept-Encoding gzip, br
	// Cf-Visitor {"scheme":"https"}
	// Accept */*
	// Paypal-Transmission-Time 2025-03-02T23:39:47Z
	// Content-Type application/json
	// X-Forwarded-Proto https
	// Connection close
	// X-Forwarded-For 173.0.81.65, 172.69.34.129
	// Cf-Ipcountry US
	// User-Agent PayPal/AUHD-214.0-58843824
	// Cf-Connecting-Ip 173.0.81.65
	// 173.0.81.65
	// /api/billing/paypal_webhook
	// {
	// 	"id": "WH-77687562XN25889J8-8Y6T55435R66168T6",
	// 	"create_time": "2018-19-12T22:20:32.000Z",
	// 	"event_type": "BILLING.SUBSCRIPTION.ACTIVATED",
	// 	"event_version": "1.0",
	// 	"resource_type": "subscription",
	// 	"resource_version": "2.0",
	// 	"summary": "A billing agreement was activated.",
	// 	"resource": {
	// 		"id": "I-BW452GLLEP1G",
	// 		"status": "ACTIVE",
	// 		"status_update_time": "2018-12-10T21:20:49Z",
	// 		"plan_id": "P-5ML4271244454362WXNWU5NQ",
	// 		"start_time": "2018-11-01T00:00:00Z",
	// 		"quantity": "20",
	// 		"shipping_amount": {
	// 			"currency_code": "USD",
	// 			"value": "10.00"
	// 		},
	// 		"subscriber": {
	// 			"name": {
	// 				"given_name": "John",
	// 				"surname": "Doe"
	// 			},
	// 			"email_address": "customer@example.com",
	// 			"shipping_address": {
	// 				"name": {
	// 					"full_name": "John Doe"
	// 				},
	// 				"address": {
	// 					"address_line_1": "2211 N First Street",
	// 					"address_line_2": "Building 17",
	// 					"admin_area_2": "San Jose",
	// 					"admin_area_1": "CA",
	// 					"postal_code": "95131",
	// 					"country_code": "US"
	// 				}
	// 			}
	// 		},
	// 		"auto_renewal": true,
	// 		"billing_info": {
	// 			"outstanding_balance": {
	// 				"currency_code": "USD",
	// 				"value": "10.00"
	// 			},
	// 			"cycle_executions": [
	// 				{
	// 					"tenure_type": "TRIAL",
	// 					"sequence": 1,
	// 					"cycles_completed": 1,
	// 					"cycles_remaining": 0,
	// 					"current_pricing_scheme_version": 1
	// 				},
	// 				{
	// 					"tenure_type": "REGULAR",
	// 					"sequence": 2,
	// 					"cycles_completed": 1,
	// 					"cycles_remaining": 0,
	// 					"current_pricing_scheme_version": 2
	// 				}
	// 			],
	// 			"last_payment": {
	// 				"amount": {
	// 					"currency_code": "USD",
	// 					"value": "500.00"
	// 				},
	// 				"time": "2018-12-01T01:20:49Z"
	// 			},
	// 			"next_billing_time": "2019-01-01T00:20:49Z",
	// 			"final_payment_time": "2020-01-01T00:20:49Z",
	// 			"failed_payments_count": 2
	// 		},
	// 		"create_time": "2018-12-10T21:20:49Z",
	// 		"update_time": "2018-12-10T21:20:49Z",
	// 		"links": [
	// 			{
	// 				"href": "https://api.paypal.com/v1/billing/subscriptions/I-BW452GLLEP1G",
	// 				"rel": "self",
	// 				"method": "GET"
	// 			},
	// 			{
	// 				"href": "https://api.paypal.com/v1/billing/subscriptions/I-BW452GLLEP1G",
	// 				"rel": "edit",
	// 				"method": "PATCH"
	// 			},
	// 			{
	// 				"href": "https://api.paypal.com/v1/billing/subscriptions/I-BW452GLLEP1G/suspend",
	// 				"rel": "suspend",
	// 				"method": "POST"
	// 			},
	// 			{
	// 				"href": "https://api.paypal.com/v1/billing/subscriptions/I-BW452GLLEP1G/cancel",
	// 				"rel": "cancel",
	// 				"method": "POST"
	// 			},
	// 			{
	// 				"href": "https://api.paypal.com/v1/billing/subscriptions/I-BW452GLLEP1G/capture",
	// 				"rel": "capture",
	// 				"method": "POST"
	// 			}
	// 		]
	// 	},
	// 	"links": [
	// 		{
	// 			"href": "https://api.paypal.com/v1/notifications/webhooks-events/WH-77687562XN25889J8-8Y6T55435R66168T6",
	// 			"rel": "self",
	// 			"method": "GET"
	// 		},
	// 		{
	// 			"href": "https://api.paypal.com/v1/notifications/webhooks-events/WH-77687562XN25889J8-8Y6T55435R66168T6/resend",
	// 			"rel": "resend",
	// 			"method": "POST"
	// 		}
	// 	]
	// }
}

func registerBillingRoutes(g *gin.Engine) {
	g.GET("/api/billing/skus", getSKUs)
	g.POST("/api/billing/payments", createPayment)
	g.GET("/api/billing/callback", paymentCallback)
	g.Any("/api/billing/cancelled", paymentCancelled)

	g.POST("/api/billing/paypal_webhook", paypalWebhook)

}
