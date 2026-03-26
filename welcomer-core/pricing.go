package welcomer

import (
	"os"

	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
)

type Currency string

const (
	CurrencyEUR Currency = "EUR"
	CurrencyGBP Currency = "GBP"
	CurrencyUSD Currency = "USD"
)

type SKUName string

const (
	SKUCustomBackgrounds SKUName = "WEL/CBG"
	SKUWelcomerPro       SKUName = "WEL/1P1"
	// SKUWelcomerProBiAnnual SKUName = "WEL/1P6"
	SKUWelcomerProAnnual   SKUName = "WEL/1P12"
	SKUWelcomerProLifetime SKUName = "WEL/1PLT"
)

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
	PatreonCheckoutID string                  `json:"patreon_checkout_id"`

	IsRecurring          bool                `json:"is_recurring"`
	PaypalSubscriptionID map[Currency]string `json:"paypal_subscription_id"`
}

var SKUPricingTable = map[int]map[SKUName]PricingSKU{
	0: {
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
				CurrencyEUR: "5.00",
				CurrencyGBP: "4.00",
				CurrencyUSD: "5.00",
			},
			IsRecurring: os.Getenv("WELCOMER_PRO_RECURRING") == "true",
			PaypalSubscriptionID: map[Currency]string{
				CurrencyEUR: os.Getenv("WELCOMER_PRO_PAYPAL_SUBSCRIPTION_EUR_ID"),
				CurrencyGBP: os.Getenv("WELCOMER_PRO_PAYPAL_SUBSCRIPTION_GBP_ID"),
				CurrencyUSD: os.Getenv("WELCOMER_PRO_PAYPAL_SUBSCRIPTION_USD_ID"),
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
				CurrencyEUR: "50.00",
				CurrencyGBP: "40.00",
				CurrencyUSD: "50.00",
			},
		},
		SKUWelcomerProLifetime: {
			ID:             SKUWelcomerProLifetime,
			Name:           "Welcomer Pro Lifetime",
			Description:    "Unlock all Welcomer Pro features for your server, forever.",
			MembershipType: database.MembershipTypeWelcomerPro,
			SoftDescriptor: "Pro Lifetime",
			MonthCount:     -1,
			Costs: map[Currency]string{
				CurrencyEUR: "100.00",
				CurrencyGBP: "80.00",
				CurrencyUSD: "100.00",
			},
		},
	},
}
