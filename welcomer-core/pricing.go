package welcomer

import (
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type Currency string

const (
	CurrencyEUR Currency = "EUR"
	CurrencyGBP Currency = "GBP"
	CurrencyINR Currency = "INR"
	CurrencyUSD Currency = "USD"
)

type SKUName string

const (
	SKUCustomBackgrounds   SKUName = "WEL/CBG"
	SKUWelcomerPro         SKUName = "WEL/1P1"
	SKUWelcomerProBiAnnual SKUName = "WEL/1P6"
	SKUWelcomerProAnnual   SKUName = "WEL/1P12"
)

var CountryMapping = map[string]Currency{
	"IN": CurrencyINR,
}

var GlobalCurrencies = []Currency{CurrencyEUR, CurrencyGBP, CurrencyUSD}

// SKU Pricing

type PricingSKU struct {
	ID                SKUName                 `json:"id"`
	Name              string                  `json:"name"`
	Description       string                  `json:"-"`
	MembershipType    database.MembershipType `json:"-"`
	SoftDescriptor    string                  `json:"-"` // This should be 13 characters or less.
	MonthCount        int                     `json:"month_count"`
	Costs             map[Currency]string     `json:"costs"`
	PatreonCheckoutId string                  `json:"patreon_checkout_id"`
}

var SKUPricing = map[SKUName]PricingSKU{
	SKUCustomBackgrounds: {
		ID:             SKUCustomBackgrounds,
		Name:           "Custom Backgrounds",
		Description:    "One-time purchase to unlock custom welcome backgrounds for your server.",
		MembershipType: database.MembershipTypeCustomBackgrounds,
		SoftDescriptor: "Backgrounds",
		MonthCount:     -1,
		Costs: map[Currency]string{
			CurrencyEUR: "12.00",
			CurrencyGBP: "10.00",
			CurrencyUSD: "12.00",
			CurrencyINR: "300",
		},
	},
	SKUWelcomerPro: {
		ID:             SKUWelcomerPro,
		Name:           "Welcomer Pro",
		Description:    "Unlock all Welcomer Pro features for your server.",
		MembershipType: database.MembershipTypeWelcomerPro,
		SoftDescriptor: "Pro",
		MonthCount:     1,
		Costs: map[Currency]string{
			CurrencyEUR: "8.00",
			CurrencyGBP: "7.00",
			CurrencyUSD: "8.00",
			CurrencyINR: "300",
		},
	},
	SKUWelcomerProBiAnnual: {
		ID:             SKUWelcomerProBiAnnual,
		Name:           "Welcomer Pro",
		Description:    "Unlock all Welcomer Pro features for your server.",
		MembershipType: database.MembershipTypeWelcomerPro,
		SoftDescriptor: "Pro",
		MonthCount:     6,
		Costs: map[Currency]string{
			CurrencyEUR: "40.00",
			CurrencyGBP: "35.00",
			CurrencyUSD: "40.00",
			CurrencyINR: "1500",
		},
	},
	SKUWelcomerProAnnual: {
		ID:             SKUWelcomerProAnnual,
		Name:           "Welcomer Pro",
		Description:    "Unlock all Welcomer Pro features for your server.",
		MembershipType: database.MembershipTypeWelcomerPro,
		SoftDescriptor: "Pro",
		MonthCount:     12,
		Costs: map[Currency]string{
			CurrencyEUR: "80.00",
			CurrencyGBP: "70.00",
			CurrencyUSD: "80.00",
			CurrencyINR: "3000",
		},
	},
}
