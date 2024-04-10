package backend

import (
	"net"
	"net/http"

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/gin-gonic/gin"
	"github.com/plutov/paypal/v4"
)

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

type GetSKUsResponse struct {
	AvailableCurrencies []welcomer.Currency `json:"available_currencies"`
	DefaultCurrency     welcomer.Currency   `json:"default_currency"`
	SKUs                []welcomer.PricingSKU
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
		DefaultCurrency:     welcomer.DefaultCurrency,
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

		// If no currency is specified, use the default currency.
		if request.Currency == "" {
			request.Currency = welcomer.DefaultCurrency
		} else if !welcomer.SliceContains(welcomer.GlobalCurrencies, request.Currency) {
			// If a currency is not in the global currencies, check against if they are in
			// the list of other available currencies.
			response, err := backend.IPChecker.CheckIP(ctx.ClientIP(), welcomer.IPIntelFlagDynamicBanListDynamicChecks, welcomer.IPIntelOFlagShowCountry)
			if err != nil {
				backend.Logger.Warn().Err(err).IPAddr("ip", net.IP(ctx.ClientIP())).Msg("Failed to validate IP via IPIntel")
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
		userTransaction, err := backend.Database.CreateUserTransaction(backend.ctx, &database.CreateUserTransactionParams{
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
		order, err := backend.PaypalClient.CreateOrder(backend.ctx, paypal.OrderIntentCapture, []paypal.PurchaseUnitRequest{purchaseUnit}, nil, &paypal.ApplicationContext{
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

		_, err = backend.Database.UpdateUserTransaction(backend.ctx, &database.UpdateUserTransactionParams{
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

// << 302

// Route POST /api/billing/callback?token=...&PayerID=...
// /api/billing/callback?token=83H56682NP651951M&PayerID=CT46WL8N7YNGE
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

		authorizeResponse, err := backend.PaypalClient.AuthorizeOrder(backend.ctx, token, paypal.AuthorizeOrderRequest{})
		if err != nil {
			backend.Logger.Error().Err(err).Str("token", token).Msg("Failed to authorize order")

			ctx.JSON(http.StatusInternalServerError, BaseResponse{
				Ok: false,
			})

			return
		}

		// Capture the order.
		// Create completed user transaction.
		// Create user membership.
		// Send user discord message.

		ctx.Header("Location", "https://"+backend.Options.Domain+"/premium#success")
		ctx.Status(http.StatusTemporaryRedirect)
	})
}

func registerBillingRoutes(g *gin.Engine) {
	g.GET("/api/billing/skus", getSKUs)
	g.POST("/api/billing/payments", createPayment)
	g.Any("/api/billing/cancelled", paymentCancelled)
}
