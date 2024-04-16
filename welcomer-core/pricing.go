package welcomer

import (
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

var DefaultCurrency = CurrencyGBP

var GlobalCurrencies = []Currency{CurrencyGBP, CurrencyUSD}

var SKUPricing = map[SKUName]PricingSKU{
	SKUWelcomerPro: {
		ID:             SKUWelcomerPro,
		Name:           "Welcomer Pro",
		Description:    "",
		MembershipType: database.MembershipTypeWelcomerPro,
		SoftDescriptor: "Pro",
		MonthCount:     1,
		Costs: map[Currency]string{
			CurrencyEUR: "7.99",
			CurrencyGBP: "6.99",
			CurrencyUSD: "7.99",
			CurrencyINR: "300",
		},
	},
	SKUCustomBackgrounds: {
		ID:             SKUCustomBackgrounds,
		Name:           "Custom Backgrounds",
		Description:    "",
		MembershipType: database.MembershipTypeCustomBackgrounds,
		SoftDescriptor: "Backgrounds",
		MonthCount:     -1,
		Costs: map[Currency]string{
			CurrencyEUR: "11.99",
			CurrencyGBP: "9.99",
			CurrencyUSD: "11.99",
			CurrencyINR: "300",
		},
	},
}

type PricingSKU struct {
	ID             SKUName                 `json:"id"`
	Name           string                  `json:"name"`
	Description    string                  `json:"-"`
	MembershipType database.MembershipType `json:"-"`
	SoftDescriptor string                  `json:"-"` // This should be 13 characters or less.
	MonthCount     int                     `json:"months"`
	Costs          map[Currency]string     `json:"costs"`
}

// Plan Pricing

type SKUName string

type Currency string

const (
	CurrencyEUR Currency = "EUR"
	CurrencyGBP Currency = "GBP"
	CurrencyINR Currency = "INR"
	CurrencyUSD Currency = "USD"
)

const (
	SKULegacyCustomBackgrounds SKUName = "WEL/CBGL"
	SKULegacyWelcomerPro1      SKUName = "WEL/1L"
	SKULegacyWelcomerPro3      SKUName = "WEL/3L"
	SKULegacyWelcomerPro5      SKUName = "WEL/5L"
	SKUWelcomerPro             SKUName = "WEL/1"
	SKUCustomBackgrounds       SKUName = "WEL/CBG"
)

var CountryMapping = map[string]Currency{
	"IN": CurrencyINR,
}
