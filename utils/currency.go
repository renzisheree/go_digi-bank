package utils

const (
	USD = "USD"
	EUR = "EUR"
	VND = "VND"
)

func IsValidCurrency(currency string) bool {
	switch currency {
	case USD, EUR, VND:
		return true
	}
	return false
}
