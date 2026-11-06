package money

import (
	"fmt"
	"strings"
	"sync"
)

// Currency is an ISO 4217 currency and the number of digits in its minor unit.
type Currency struct {
	Code   string
	Digits int
}

var (
	registryMu sync.RWMutex
	registry   = map[string]int{
		"AED": 2, "ARS": 2, "AUD": 2, "BDT": 2, "BGN": 2, "BHD": 3, "BRL": 2,
		"CAD": 2, "CHF": 2, "CLP": 0, "CNY": 2, "COP": 2, "CZK": 2, "DKK": 2,
		"EGP": 2, "EUR": 2, "GBP": 2, "GHS": 2, "HKD": 2, "HUF": 2, "IDR": 2,
		"ILS": 2, "INR": 2, "ISK": 0, "JOD": 3, "JPY": 0, "KES": 2, "KRW": 0,
		"KWD": 3, "LKR": 2, "MAD": 2, "MXN": 2, "MYR": 2, "NGN": 2, "NOK": 2,
		"NPR": 2, "NZD": 2, "OMR": 3, "PHP": 2, "PKR": 2, "PLN": 2, "QAR": 2,
		"RON": 2, "SAR": 2, "SEK": 2, "SGD": 2, "THB": 2, "TND": 3, "TRY": 2,
		"TWD": 2, "UAH": 2, "UGX": 0, "USD": 2, "VND": 0, "XAF": 0, "XOF": 0,
		"ZAR": 2,
	}
)

// Lookup returns the registered currency for code. Codes are case-insensitive.
func Lookup(code string) (Currency, error) {
	code = strings.ToUpper(code)

	registryMu.RLock()
	digits, ok := registry[code]
	registryMu.RUnlock()

	if !ok {
		return Currency{}, fmt.Errorf("%w: %q", ErrUnknownCurrency, code)
	}
	return Currency{Code: code, Digits: digits}, nil
}

// Register adds a currency or replaces the digits of an existing one.
func Register(code string, digits int) error {
	if len(code) != 3 || strings.ToUpper(code) != code || strings.ContainsFunc(code, func(r rune) bool { return r < 'A' || r > 'Z' }) {
		return fmt.Errorf("money: invalid currency code %q", code)
	}
	if digits < 0 || digits > 8 {
		return fmt.Errorf("money: digits must be between 0 and 8, got %d", digits)
	}

	registryMu.Lock()
	registry[code] = digits
	registryMu.Unlock()
	return nil
}
