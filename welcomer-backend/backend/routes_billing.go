package backend

import (
	"net"
	"net/http"

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/gin-gonic/gin"
	"github.com/plutov/paypal/v4"
)

const defaultCurrency = CurrencyGBP

var globalCurrencies = []Currency{CurrencyGBP, CurrencyUSD}

var skuPricing = map[SKUName]PricingSKU{
	SKUWelcomerPro: {
		ID:             SKUWelcomerPro,
		Name:           "Welcomer Pro",
		Description:    "",
		SoftDescriptor: "Welcomer Pro",
		Costs: map[Currency]string{
			CurrencyGBP: "5.00",
			CurrencyUSD: "5.00",
			CurrencyINR: "300",
		},
	},
	SKUCustomBackgrounds: {
		ID:             SKUCustomBackgrounds,
		Name:           "Custom Backgrounds",
		Description:    "",
		SoftDescriptor: "Custom Bgs",
		Costs: map[Currency]string{
			CurrencyGBP: "8.00",
			CurrencyUSD: "8.00",
			CurrencyINR: "400",
		},
	},
}

// Plan Pricing

type SKUName string

type Currency string

const (
	CurrencyGBP Currency = "GBP"
	CurrencyUSD Currency = "USD"
	CurrencyINR Currency = "INR"
)

const (
	SKULegacyCustomBackgrounds SKUName = "WEL/CBGL"
	SKULegacyWelcomerPro1      SKUName = "WEL/1L"
	SKULegacyWelcomerPro3      SKUName = "WEL/3L"
	SKULegacyWelcomerPro5      SKUName = "WEL/5L"
	SKUWelcomerPro             SKUName = "WEL/1"
	SKUCustomBackgrounds       SKUName = "WEL/CBG"
)

var countryMapping = map[string]Currency{
	"IN": CurrencyINR,
}

type Pricing struct {
	AvailableCurrencies []Currency   `json:"available_currencies"`
	DefaultCurrency     Currency     `json:"default_currency"`
	SKUs                []PricingSKU `json:"skus"`
}

type PricingSKU struct {
	ID             SKUName                 `json:"id"`
	Name           string                  `json:"name"`
	Description    string                  `json:"-"`
	MembershipType database.MembershipType `json:"-"`
	SoftDescriptor string                  `json:"-"` // This should be 13 characters or less.
	Costs          map[Currency]string     `json:"costs"`
}

func getAvailableCurrencies(ipintelResponse welcomer.IPIntelResponse) []Currency {
	// If the IPIntel response is above the threshold, we assume the user is on a VPN.
	if ipintelResponse.Result <= IPIntelThreshold {
		mapping, ok := countryMapping[ipintelResponse.Country]
		if ok {
			return append(globalCurrencies, mapping)
		}
	}

	return globalCurrencies
}

type BillingSKUs struct {
	AvailableCurrencies []Currency `json:"available_currencies"`
	DefaultCurrency     Currency   `json:"default_currency"`
	SKUs                []PricingSKU
}

// Route GET /api/billing/skus
func getSKUs(ctx *gin.Context) {
	response, err := backend.IPChecker.CheckIP(ctx.ClientIP(), welcomer.IPIntelFlagDynamicBanListDynamicChecks, welcomer.IPIntelOFlagShowCountry)
	if err != nil {
		backend.Logger.Warn().Err(err).IPAddr("ip", net.IP(ctx.ClientIP())).Msg("Failed to validate IP via IPIntel")
	}

	currencies := getAvailableCurrencies(response)

	pricingStructure := BillingSKUs{
		AvailableCurrencies: currencies,
		DefaultCurrency:     defaultCurrency,
		SKUs:                make([]PricingSKU, 0, len(skuPricing)),
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
	SKU      SKUName  `json:"sku"`
	Currency Currency `json:"currency"`
}

type CreatePaymentResponse struct {
	URL string `json:"url"`
}

// Route POST /api/billing/payments
func createPayment(ctx *gin.Context) {
	// requireOAuthAuthorization(ctx, func(ctx *gin.Context) {
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

	sku, ok := skuPricing[SKUName(request.SKU)]
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
		request.Currency = defaultCurrency
	} else if !welcomer.SliceContains(globalCurrencies, request.Currency) {
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
	// })
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
// << 302
// /api/billing/callback?token=83H56682NP651951M&PayerID=CT46WL8N7YNGE

func registerBillingRoutes(g *gin.Engine) {
	g.GET("/api/billing/skus", getSKUs)
	g.POST("/api/billing/payments", createPayment)
	g.Any("/api/billing/cancelled", paymentCancelled)
}
