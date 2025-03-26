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
	var response utils.IPIntelResponse

	var err error

	if os.Getenv("FT_USE_CLOUDFLARE_IPCOUNTRY") == "true" {
		response = utils.IPIntelResponse{
			Result:  0,
			Country: ctx.GetHeader("CF-IPCountry"),
		}
	} else {
		response, err = backend.IPChecker.CheckIP(ctx, ctx.ClientIP(), utils.IPIntelFlagDynamicBanListDynamicChecks, utils.IPIntelOFlagShowCountry)
		if err != nil {
			backend.Logger.Warn().Err(err).IPAddr("ip", net.IP(ctx.ClientIP())).Msg("Failed to validate IP via IPIntel")
		}
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
				applicationContext.ReturnURL = "https://" + backend.Options.Domain + "/api/billing/subscription_callback"

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

// Route POST /api/billing/subscription_callback?
func paymentSubscriptionCallback(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		err := backend.Pool.BeginTxFunc(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
			// https://beta.welcomer.gg/api/billing/callback?subscription_id=I-TMR758P4PW2C&ba_token=BA-1UW93399KE424124U&token=6YC07506GL255092S

			queries := backend.Database.WithTx(tx)

			// Read "subscription_id" from the query string.
			subscriptionID := ctx.Query("subscription_id")

			if subscriptionID == "" {
				backend.Logger.Warn().Msg("Missing subscription_id")

				ctx.JSON(http.StatusBadRequest, NewBaseResponse(NewMissingParameterError("subscription_id"), nil))

				return ErrMissingParameter
			}

			transactions, err := queries.GetUserTransactionsByTransactionID(ctx, subscriptionID)
			if err != nil {
				backend.Logger.Error().Err(err).Str("subscription_id", subscriptionID).Msg("Failed to get user transactions by transaction ID")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

				return err
			}

			if len(transactions) == 0 {
				backend.Logger.Warn().Str("subscription_id", subscriptionID).Msg("No user transactions found")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

				return err
			}

			user := tryGetUser(ctx)
			transaction := transactions[0]

			if transaction.UserID != int64(user.ID) {
				backend.Logger.Warn().Str("subscription_id", subscriptionID).Int64("userID", transaction.UserID).Int64("user.ID", int64(user.ID)).Msg("User ID does not match")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

				return err
			}

			// TODO: If we receive the webhook before the user navigates
			// to the success page, it may be completed.
			if transaction.TransactionStatus != int32(database.TransactionStatusPending) {
				backend.Logger.Warn().Str("subscription_id", subscriptionID).Str("transactionStatus", database.TransactionStatus(transaction.TransactionStatus).String()).Msg("Transaction is not pending")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

				return err
			}

			// Get subscription
			subscription, err := backend.PaypalClient.GetSubscriptionDetails(ctx, subscriptionID)
			if err != nil {
				backend.Logger.Error().Err(err).Str("subscription_id", subscriptionID).Msg("Failed to get subscription")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(ErrGetOrderValidationFailed, nil))

				return err
			}

			// Create a paypal subscription entry.
			_, err = queries.CreateOrUpdatePaypalSubscription(ctx, database.CreateOrUpdatePaypalSubscriptionParams{
				SubscriptionID:     subscriptionID,
				UserID:             int64(user.ID),
				PayerID:            subscription.Subscriber.PayerID,
				LastBilledAt:       subscription.BillingInfo.LastPayment.Time,
				NextBillingAt:      subscription.BillingInfo.NextBillingTime,
				SubscriptionStatus: string(subscription.SubscriptionStatus),
				PlanID:             subscription.PlanID,
				Quantity:           subscription.Quantity,
			})
			if err != nil {
				backend.Logger.Error().Err(err).
					Str("subscription_id", subscription.ID).
					Msg("Failed to create paypal subscription")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

				return err
			}

			// Fetch SKU from the plan ID.
			pricing := getSKUPricing()

			var sku welcomer.PricingSKU

			for _, pricingSKU := range pricing {
				if pricingSKU.IsRecurring {
					for _, plan := range pricingSKU.PaypalSubscriptionID {
						if plan == subscription.PlanID {
							sku = pricingSKU

							break
						}
					}
				}
			}

			if sku.ID == "" {
				backend.Logger.Warn().Str("planID", subscription.PlanID).Msg("Invalid plan ID")

				ctx.JSON(http.StatusBadRequest, NewBaseResponse(ErrInvalidSKU, nil))

				return err
			}

			// Create a user transaction.
			userTransaction, err := welcomer.CreateTransactionForUser(
				ctx,
				queries,
				user.ID,
				database.PlatformTypePaypalSubscription,
				database.TransactionStatusPending,
				subscriptionID,
				transaction.CurrencyCode,
				transaction.Amount,
			)
			if err != nil {
				backend.Logger.Error().Err(err).Msg("Failed to create user transaction")

				ctx.JSON(http.StatusInternalServerError, NewBaseResponse(ErrCreateTransactionFailed, nil))

				return err
			}

			startedAt := time.Now()
			expiresAt := startedAt.AddDate(0, 0, 7)

			// Create a temporary membership until we receive
			// a payment confirmation from paypal.
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
	message := fmt.Sprintf("%s|%s|%s|%d", transmissionID, timeStamp, os.Getenv("PAYPAL_WEBHOOK_ID"), crc)

	certPemBytes, err := downloadAndCache(certURL)
	if err != nil {
		backend.Logger.Error().Err(err).Msg("Failed to download and cache certificate")

		return event, false
	}

	block, _ := pem.Decode(certPemBytes)
	if block == nil {
		backend.Logger.Error().Msg("Failed to decode PEM block containing the certificate")

		return event, false
	}

	certPem, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		backend.Logger.Error().Err(err).Msg("Failed to decode certificate")

		return event, false
	}

	signature, err := base64.StdEncoding.DecodeString(headers.Get("Paypal-Transmission-Sig"))
	if err != nil {
		backend.Logger.Error().Err(err).Msg("Failed to decode signature")

		return event, false
	}

	hashed := crypto.SHA256.New()
	hashed.Write([]byte(message))

	pub, ok := certPem.PublicKey.(*rsa.PublicKey)
	if !ok {
		backend.Logger.Error().Msg("Failed to get public key")

		return event, false
	}

	err = rsa.VerifyPKCS1v15(pub, crypto.SHA256, hashed.Sum(nil), signature)
	if err != nil {
		backend.Logger.Error().Msg("Failed to verify signature")

		return event, false
	}

	return event, true
}

type PaypalWebhookEvent string

const (
	WebhookEventPaymentSaleCompleted             PaypalWebhookEvent = "PAYMENT.SALE.COMPLETED"
	WebhookEventBillingSubscriptionCreated       PaypalWebhookEvent = "BILLING.SUBSCRIPTION.CREATED"
	WebhookEventBillingSubscriptionActivated     PaypalWebhookEvent = "BILLING.SUBSCRIPTION.ACTIVATED"
	WebhookEventBillingSubscriptionUpdated       PaypalWebhookEvent = "BILLING.SUBSCRIPTION.UPDATED"
	WebhookEventBillingSubscriptionReactivated   PaypalWebhookEvent = "BILLING.SUBSCRIPTION.RE-ACTIVATED"
	WebhookEventBillingSubscriptionExpired       PaypalWebhookEvent = "BILLING.SUBSCRIPTION.EXPIRED"
	WebhookEventBillingSubscriptionCancelled     PaypalWebhookEvent = "BILLING.SUBSCRIPTION.CANCELLED"
	WebhookEventBillingSubscriptionSuspended     PaypalWebhookEvent = "BILLING.SUBSCRIPTION.SUSPENDED"
	WebhookEventBillingSubscriptionPaymentFailed PaypalWebhookEvent = "BILLING.SUBSCRIPTION.PAYMENT.FAILED"
)

// Route POST /api/billing/paypal_webhook
func paypalWebhook(ctx *gin.Context) {
	event, ok := validatePaypalWebhook(ctx)
	if !ok {
		// ctx.JSON(http.StatusForbidden, NewBaseResponse(ErrWebhookValidationFailed, nil))

		// return
	}

	// Handle different event types
	var webhookEvent paypal.AnyEvent
	if err := json.Unmarshal(event, &webhookEvent); err != nil {
		backend.Logger.Error().Err(err).Msg("Failed to unmarshal webhook event")

		ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

		return
	}

	if PaypalWebhookEvent(webhookEvent.EventType) == WebhookEventPaymentSaleCompleted {
		var sale welcomer.PaypalSale

		if err := json.Unmarshal(webhookEvent.Resource, &sale); err != nil {
			backend.Logger.Error().Err(err).Msg("Failed to unmarshal resource ID")

			ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

			return
		}

		err := welcomer.HandlePaypalSale(ctx, backend.Logger, backend.Database, sale)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, NewBaseResponse(err, nil))

			return
		}

		ctx.JSON(http.StatusOK, NewBaseResponse(nil, nil))
	}

	var subscription paypal.Subscription

	if err := json.Unmarshal(webhookEvent.Resource, &subscription); err != nil {
		backend.Logger.Error().Err(err).Msg("Failed to unmarshal subscription")

		ctx.JSON(http.StatusInternalServerError, NewBaseResponse(NewGenericErrorWithLineNumber(), nil))

		return
	}

	var err error

	switch PaypalWebhookEvent(webhookEvent.EventType) {
	case WebhookEventBillingSubscriptionCreated:
		err = welcomer.OnPaypalSubscriptionCreated(ctx, backend.Logger, backend.Database, subscription)
		if err != nil {
			backend.Logger.Error().Err(err).Msg("Failed to handle subscription created event")
		}

	case WebhookEventBillingSubscriptionActivated:
		err = welcomer.OnPaypalSubscriptionActivated(ctx, backend.Logger, backend.Database, subscription)
		if err != nil {
			backend.Logger.Error().Err(err).Msg("Failed to handle subscription activated event")
		}

	case WebhookEventBillingSubscriptionUpdated:
		err = welcomer.OnPaypalSubscriptionUpdated(ctx, backend.Logger, backend.Database, subscription)
		if err != nil {
			backend.Logger.Error().Err(err).Msg("Failed to handle subscription updated event")
		}

	case WebhookEventBillingSubscriptionReactivated:
		err = welcomer.OnPaypalSubscriptionReactivated(ctx, backend.Logger, backend.Database, subscription)
		if err != nil {
			backend.Logger.Error().Err(err).Msg("Failed to handle subscription re-activated event")
		}

	case WebhookEventBillingSubscriptionExpired:
		err = welcomer.OnPaypalSubscriptionExpired(ctx, backend.Logger, backend.Database, subscription)
		if err != nil {
			backend.Logger.Error().Err(err).Msg("Failed to handle subscription expired event")
		}

	case WebhookEventBillingSubscriptionCancelled:
		err = welcomer.OnPaypalSubscriptionCancelled(ctx, backend.Logger, backend.Database, subscription)
		if err != nil {
			backend.Logger.Error().Err(err).Msg("Failed to handle subscription cancelled event")
		}

	case WebhookEventBillingSubscriptionSuspended:
		err = welcomer.OnPaypalSubscriptionSuspended(ctx, backend.Logger, backend.Database, subscription)
		if err != nil {
			backend.Logger.Error().Err(err).Msg("Failed to handle subscription suspended event")
		}

	case WebhookEventBillingSubscriptionPaymentFailed:
		err = welcomer.OnPaypalSubscriptionPaymentFailed(ctx, backend.Logger, backend.Database, subscription)
		if err != nil {
			backend.Logger.Error().Err(err).Msg("Failed to handle subscription payment failed event")
		}

	default:
		backend.Logger.Warn().
			Str("event_type", webhookEvent.EventType).
			RawJSON("data", webhookEvent.Resource).
			Msg("Unhandled event type")
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, NewBaseResponse(err, nil))

		return
	}

	ctx.JSON(http.StatusOK, NewBaseResponse(nil, nil))
}

func registerBillingRoutes(g *gin.Engine) {
	g.GET("/api/billing/skus", getSKUs)
	g.POST("/api/billing/payments", createPayment)
	g.GET("/api/billing/callback", paymentCallback)
	g.GET("/api/billing/subscription_callback", paymentSubscriptionCallback)
	g.Any("/api/billing/cancelled", paymentCancelled)

	g.POST("/api/billing/paypal_webhook", paypalWebhook)
}
