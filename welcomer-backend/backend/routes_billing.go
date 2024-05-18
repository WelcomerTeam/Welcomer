package backend

import (
	"net"
	"net/http"
	"time"

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/gin-gonic/gin"
	"github.com/plutov/paypal/v4"
)

// ISO 3166-1 alpha-2 country codes for the Eurozone.
var euroZone = []string{"AT", "BE", "HR", "CY", "EE", "FI", "FR", "DE", "GR", "IE", "IT", "LV", "LT", "LU", "MT", "NL", "PT", "SK", "SI", "ES"}

func getAvailableCurrencies(ipintelResponse welcomer.IPIntelResponse) []welcomer.Currency {
	// If the IPIntel response is above the threshold, we assume the user is on a VPN.
	if ipintelResponse.Result <= IPIntelThreshold {
		mapping, ok := welcomer.CountryMapping[ipintelResponse.Country]
		if ok {
			return append(welcomer.GlobalCurrencies, mapping)
		}
	}

	return welcomer.GlobalCurrencies
}

func getDefaultCurrency(ipIntelResponse welcomer.IPIntelResponse) welcomer.Currency {
	if ipIntelResponse.Country == "GB" {
		return welcomer.CurrencyGBP
	} else if ipIntelResponse.Country == "IN" {
		return welcomer.CurrencyINR
	} else if welcomer.SliceContains(euroZone, ipIntelResponse.Country) {
		return welcomer.CurrencyEUR
	}

	return welcomer.CurrencyUSD
}

type GetSKUsResponse struct {
	AvailableCurrencies []welcomer.Currency   `json:"available_currencies"`
	DefaultCurrency     welcomer.Currency     `json:"default_currency"`
	SKUs                []welcomer.PricingSKU `json:"skus"`
}

// Route GET /api/billing/skus
func getSKUs(ctx *gin.Context) {
	response, err := backend.IPChecker.CheckIP(ctx.ClientIP(), welcomer.IPIntelFlagDynamicBanListDynamicChecks, welcomer.IPIntelOFlagShowCountry)
	if err != nil {
		backend.Logger.Warn().Err(err).IPAddr("ip", net.IP(ctx.ClientIP())).Msg("Failed to validate IP via IPIntel")
	}

	currencies := getAvailableCurrencies(response)

	pricingStructure := GetSKUsResponse{
		AvailableCurrencies: currencies,
		DefaultCurrency:     getDefaultCurrency(response),
		SKUs:                make([]welcomer.PricingSKU, 0, len(welcomer.SKUPricing)),
	}

	for _, sku := range welcomer.SKUPricing {
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

// Route POST /api/billing/payments
func createPayment(ctx *gin.Context) {
	requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
		// Read "request" text from the post json body.
		var request CreatePaymentRequest

		if err := ctx.ShouldBindJSON(&request); err != nil {
			backend.Logger.Warn().Err(err).Msg("Failed to bind JSON")

			ctx.JSON(http.StatusBadRequest, BaseResponse{
				Ok: false,
			})

			return
		}

		if request.SKU == "" {
			ctx.JSON(http.StatusBadRequest, BaseResponse{
				Ok:    false,
				Error: "missing sku",
			})

			return
		}

		sku, ok := welcomer.SKUPricing[welcomer.SKUName(request.SKU)]
		if !ok {
			backend.Logger.Warn().Str("sku", string(request.SKU)).Msg("Invalid SKU")

			ctx.JSON(http.StatusBadRequest, BaseResponse{
				Ok:    false,
				Error: "invalid sku",
			})

			return
		}

		response, err := backend.IPChecker.CheckIP(ctx.ClientIP(), welcomer.IPIntelFlagDynamicBanListDynamicChecks, welcomer.IPIntelOFlagShowCountry)
		if err != nil {
			backend.Logger.Warn().Err(err).IPAddr("ip", net.IP(ctx.ClientIP())).Msg("Failed to validate IP via IPIntel")
		}

		if request.Currency == "" {
			request.Currency = getDefaultCurrency(response)
		}

		currencies := getAvailableCurrencies(response)
		if !welcomer.SliceContains(currencies, request.Currency) {
			backend.Logger.Warn().Str("currency", string(request.Currency)).Msg("Invalid currency")

			ctx.JSON(http.StatusBadRequest, BaseResponse{
				Ok:    false,
				Error: "currency specified is not available for this user",
			})

			return
		}

		// Check if the currency is available for this SKU.
		skuCost, ok := sku.Costs[request.Currency]
		if !ok {
			backend.Logger.Warn().Str("currency", string(request.Currency)).Msg("Invalid currency")

			ctx.JSON(http.StatusBadRequest, BaseResponse{
				Ok:    false,
				Error: "currency specified is not available for this sku",
			})

			return
		}

		// Check if the cost is valid.
		if welcomer.TryParseFloat(skuCost) <= 0 {
			backend.Logger.Warn().Str("currency", string(request.Currency)).Str("sku", string(request.SKU)).Msg("Invalid cost")

			ctx.JSON(http.StatusInternalServerError, BaseResponse{
				Ok: false,
			})

			return
		}

		user := tryGetUser(ctx)

		// Create a transaction for the user.
		money := &paypal.Money{
			Currency: string(request.Currency),
			Value:    skuCost,
		}

		// Create a user transaction.
		userTransaction, err := backend.Database.CreateUserTransaction(backend.ctx, database.CreateUserTransactionParams{
			UserID:            int64(user.ID),
			PlatformType:      int32(database.PlatformTypePaypal),
			TransactionID:     "",
			TransactionStatus: int32(database.TransactionStatusPending),
			CurrencyCode:      money.Currency,
			Amount:            money.Value,
		})
		if err != nil {
			backend.Logger.Error().Err(err).Msg("Failed to create user transaction")

			ctx.JSON(http.StatusInternalServerError, BaseResponse{
				Ok: false,
			})

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
					SKU:         string(request.SKU),
					Category:    paypal.ItemCategoryDigitalGood,
				},
			},
			Shipping:           nil,
			PaymentInstruction: nil,
		}

		// Send order request to paypal.
		order, err := backend.PaypalClient.CreateOrder(backend.ctx, paypal.OrderIntentAuthorize, []paypal.PurchaseUnitRequest{purchaseUnit}, nil, &paypal.ApplicationContext{
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
		})
		if err != nil {
			backend.Logger.Error().Err(err).Msg("Failed to create order")

			ctx.JSON(http.StatusInternalServerError, BaseResponse{
				Ok: false,
			})

			return
		}

		backend.Logger.Info().
			Str("orderID", order.ID).
			Str("transactionUuid", userTransaction.TransactionUuid.String()).
			Msg("Created order")

		// Update the user transaction with the order ID.
		userTransaction.TransactionID = order.ID

		_, err = backend.Database.UpdateUserTransaction(backend.ctx, database.UpdateUserTransactionParams{
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

			ctx.JSON(http.StatusInternalServerError, BaseResponse{
				Ok: false,
			})

			return
		}

		// Get the payer action link.
		payerActionLink := getLinkNamed(order.Links, "approve")
		if payerActionLink == "" {
			backend.Logger.Error().Msg("Failed to get payer action link from order response")

			ctx.JSON(http.StatusInternalServerError, BaseResponse{
				Ok: false,
			})

			return
		}

		ctx.JSON(http.StatusOK, BaseResponse{
			Ok: true,
			Data: CreatePaymentResponse{
				URL: payerActionLink,
			},
		})
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

		// Read "token" and "PayerID" from the query string.
		token := ctx.Query("token")
		payerID := ctx.Query("PayerID")

		if token == "" {
			backend.Logger.Warn().Str("PayerID", payerID).Msg("Missing token")

			ctx.JSON(http.StatusBadRequest, BaseResponse{
				Ok:    false,
				Error: "missing token",
			})

			return
		}

		if payerID == "" {
			backend.Logger.Warn().Str("token", token).Msg("Missing PayerID")

			ctx.JSON(http.StatusBadRequest, BaseResponse{
				Ok:    false,
				Error: "missing PayerID",
			})

			return
		}

		transactions, err := backend.Database.GetUserTransactionsByTransactionID(backend.ctx, token)
		if err != nil {
			backend.Logger.Error().Err(err).Str("token", token).Msg("Failed to get user transactions by transaction ID")

			ctx.JSON(http.StatusInternalServerError, BaseResponse{
				Ok: false,
			})

			return
		}

		if len(transactions) == 0 {
			backend.Logger.Warn().Str("token", token).Msg("No user transactions found")

			ctx.JSON(http.StatusInternalServerError, BaseResponse{
				Ok: false,
			})

			return
		}

		user := tryGetUser(ctx)
		transaction := transactions[0]

		if transaction.UserID != int64(user.ID) {
			backend.Logger.Warn().Str("token", token).Int64("userID", transaction.UserID).Int64("user.ID", int64(user.ID)).Msg("User ID does not match")

			ctx.JSON(http.StatusInternalServerError, BaseResponse{
				Ok: false,
			})

			return
		}

		if transaction.TransactionStatus != int32(database.TransactionStatusPending) {
			backend.Logger.Warn().Str("token", token).Str("transactionStatus", database.TransactionStatus(transaction.TransactionStatus).String()).Msg("Transaction is not pending")

			ctx.JSON(http.StatusInternalServerError, BaseResponse{
				Ok: false,
			})

			return
		}

		// Get order
		order, err := backend.PaypalClient.GetOrder(backend.ctx, token)
		if err != nil {
			backend.Logger.Error().Err(err).Str("token", token).Msg("Failed to get order")

			ctx.JSON(http.StatusInternalServerError, BaseResponse{
				Ok: false,
			})

			return
		}

		if len(order.PurchaseUnits) == 0 || len(order.PurchaseUnits[0].Items) == 0 {
			backend.Logger.Warn().Str("token", token).Msg("No purchase units or items found")

			ctx.JSON(http.StatusInternalServerError, BaseResponse{
				Ok: false,
			})

			return
		}

		// Fetch SKU from the order.
		skuName := order.PurchaseUnits[0].Items[0].SKU

		sku, ok := welcomer.SKUPricing[welcomer.SKUName(skuName)]
		if !ok {
			backend.Logger.Warn().Str("sku", skuName).Msg("Invalid SKU")

			ctx.JSON(http.StatusBadRequest, BaseResponse{
				Ok:    false,
				Error: "invalid sku",
			})

			return
		}

		// Capture the order
		authorizeResponse, err := backend.PaypalClient.AuthorizeOrder(backend.ctx, token, paypal.AuthorizeOrderRequest{})
		if err != nil || authorizeResponse.Status != paypal.OrderStatusCompleted {
			backend.Logger.Error().Err(err).Str("token", token).Str("status", authorizeResponse.Status).Msg("Failed to authorize order")

			// Create a user transaction.
			_, err = backend.Database.CreateUserTransaction(backend.ctx, database.CreateUserTransactionParams{
				UserID:            int64(user.ID),
				PlatformType:      int32(database.PlatformTypePaypal),
				TransactionID:     authorizeResponse.ID,
				TransactionStatus: int32(database.TransactionStatusPending),
				CurrencyCode:      transaction.CurrencyCode,
				Amount:            transaction.Amount,
			})
			if err != nil {
				backend.Logger.Error().Err(err).Msg("Failed to create user transaction")
			}

			ctx.JSON(http.StatusInternalServerError, BaseResponse{
				Ok: false,
			})

			return
		}

		// Create a user transaction.
		userTransaction, err := backend.Database.CreateUserTransaction(backend.ctx, database.CreateUserTransactionParams{
			UserID:            int64(user.ID),
			PlatformType:      int32(database.PlatformTypePaypal),
			TransactionID:     authorizeResponse.ID,
			TransactionStatus: int32(database.TransactionStatusCompleted),
			CurrencyCode:      order.PurchaseUnits[0].Amount.Currency,
			Amount:            order.PurchaseUnits[0].Amount.Value,
		})
		if err != nil {
			backend.Logger.Error().Err(err).Msg("Failed to create user transaction")

			ctx.JSON(http.StatusInternalServerError, BaseResponse{
				Ok:    false,
				Error: "Failed to create new transaction. Please contact support.",
			})

			return
		}

		startedAt := time.Time{}

		expiresAt := startedAt
		if sku.MonthCount <= 0 {
			expiresAt = startedAt.AddDate(0, welcomer.If(sku.MonthCount < 0, 120, sku.MonthCount), 0)
		}

		// Create a new membership for the user.
		_, err = backend.Database.CreateNewMembership(backend.ctx, database.CreateNewMembershipParams{
			StartedAt:       startedAt,
			ExpiresAt:       expiresAt,
			Status:          int32(database.MembershipStatusIdle),
			MembershipType:  int32(sku.MembershipType),
			TransactionUuid: userTransaction.TransactionUuid,
			UserID:          int64(user.ID),
			GuildID:         0,
		})
		if err != nil {
			backend.Logger.Error().Err(err).Msg("Failed to create new membership")

			ctx.JSON(http.StatusInternalServerError, BaseResponse{
				Ok:    false,
				Error: "Failed to create new membership. Please contact support.",
			})

			return
		}

		ctx.Header("Location", "https://"+backend.Options.Domain+"/premium#success")
		ctx.Status(http.StatusTemporaryRedirect)
	})
}

func registerBillingRoutes(g *gin.Engine) {
	g.GET("/api/billing/skus", getSKUs)
	g.POST("/api/billing/payments", createPayment)
	g.GET("/api/billing/callback", paymentCallback)
	g.Any("/api/billing/cancelled", paymentCancelled)
}
