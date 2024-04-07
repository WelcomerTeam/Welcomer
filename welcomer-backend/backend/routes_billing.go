package backend

import (
	"net"
	"net/http"

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	"github.com/gin-gonic/gin"
)

const defaultCurrency = CurrencyGBP

var globalCurrencies = []Currency{CurrencyGBP, CurrencyUSD}

var skuPricing = map[SKUName]PricingSKU{
	SKUWelcomerPro: {
		ID:   SKUWelcomerPro,
		Name: "Welcomer Pro",
		Costs: map[Currency]float64{
			CurrencyGBP: 5.00,
			CurrencyUSD: 5.00,
			CurrencyINR: 300,
		},
	},
	SKUCustomBackgrounds: {
		ID:   SKUCustomBackgrounds,
		Name: "Custom Backgrounds",
		Costs: map[Currency]float64{
			CurrencyGBP: 8.00,
			CurrencyUSD: 8.00,
			CurrencyINR: 400,
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

var skuMapping = map[database.MembershipType]SKUName{
	database.MembershipTypeLegacyCustomBackgrounds: SKULegacyCustomBackgrounds,
	database.MembershipTypeLegacyWelcomerPro1:      SKULegacyWelcomerPro1,
	database.MembershipTypeLegacyWelcomerPro3:      SKULegacyWelcomerPro3,
	database.MembershipTypeLegacyWelcomerPro5:      SKULegacyWelcomerPro5,
	database.MembershipTypeWelcomerPro:             SKUWelcomerPro,
	database.MembershipTypeCustomBackgrounds:       SKUCustomBackgrounds,
}

var countryMapping = map[string]Currency{
	"IN": CurrencyINR,
}

type Pricing struct {
	AvailableCurrencies []Currency   `json:"available_currencies"`
	DefaultCurrency     Currency     `json:"default_currency"`
	SKUs                []PricingSKU `json:"skus"`
}

type PricingSKU struct {
	ID    SKUName
	Name  string
	Costs map[Currency]float64
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

// Route POST /api/billing/skus/:sku
// >> {
// >>     "currency": "GBP"
// >> }
// << {
// <<     "success": true,
// <<     "url": ...,
// << }

// Route POST /api/billing/callback?paymentId=...&payerId=...
// << 302

func registerBillingRoutes(g *gin.Engine) {
	g.GET("/api/billing/skus", getSKUs)
}
