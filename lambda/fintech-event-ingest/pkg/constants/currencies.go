package constants

// Currency defines the type for supported currencies.
type Currency string

// Predefined supported currencies.
const (
	CurrencyUSD Currency = "USD"
	CurrencyARS Currency = "ARS"
)

// IsSupportedCurrency checks if the given currency is supported.
func IsSupportedCurrency(c Currency) bool {
	switch c {
	case CurrencyUSD, CurrencyARS:
		return true
	default:
		return false
	}
}
