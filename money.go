// Package money provides a currency-safe amount type stored in integer minor
// units, with explicit rounding for multiplication and currency conversion.
package money

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
)

var (
	ErrUnknownCurrency  = errors.New("money: unknown currency")
	ErrCurrencyMismatch = errors.New("money: currency mismatch")
	ErrOverflow         = errors.New("money: amount out of range")
	ErrInvalidAmount    = errors.New("money: invalid amount")
)

// Money is an amount in the minor unit of a currency (pence for GBP, yen for
// JPY). The zero value has no currency and should not be used.
//
// Amounts are kept within ±math.MaxInt64 so negation never overflows.
type Money struct {
	amount   int64
	currency Currency
}

// New returns minor units of the given currency.
func New(minor int64, code string) (Money, error) {
	c, err := Lookup(code)
	if err != nil {
		return Money{}, err
	}
	if minor == math.MinInt64 {
		return Money{}, ErrOverflow
	}
	return Money{amount: minor, currency: c}, nil
}

// Parse reads a decimal string such as "12.50" or "-0.3". It rejects more
// decimal places than the currency allows rather than rounding silently.
func Parse(s, code string) (Money, error) {
	c, err := Lookup(code)
	if err != nil {
		return Money{}, err
	}

	s = strings.TrimSpace(s)
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimLeft(s, "+-")

	whole, frac, _ := strings.Cut(s, ".")
	if whole == "" && frac == "" || !isDigits(whole) || !isDigits(frac) {
		return Money{}, fmt.Errorf("%w: %q", ErrInvalidAmount, s)
	}
	if len(frac) > c.Digits {
		return Money{}, fmt.Errorf("%w: %s allows %d decimal places", ErrInvalidAmount, c.Code, c.Digits)
	}

	digits := whole + frac + strings.Repeat("0", c.Digits-len(frac))
	n, _ := new(big.Int).SetString(digits, 10)
	if neg {
		n.Neg(n)
	}

	amount, err := toInt64(n)
	if err != nil {
		return Money{}, err
	}
	return Money{amount: amount, currency: c}, nil
}

// MustParse is like Parse but panics on error. Intended for constants and tests.
func MustParse(s, code string) Money {
	m, err := Parse(s, code)
	if err != nil {
		panic(err)
	}
	return m
}

func (m Money) Amount() int64      { return m.amount }
func (m Money) Currency() Currency { return m.currency }
func (m Money) IsZero() bool       { return m.amount == 0 }

func (m Money) Sign() int {
	switch {
	case m.amount < 0:
		return -1
	case m.amount > 0:
		return 1
	}
	return 0
}

func (m Money) Neg() Money {
	m.amount = -m.amount
	return m
}

func (m Money) Abs() Money {
	if m.amount < 0 {
		return m.Neg()
	}
	return m
}

func (m Money) Add(o Money) (Money, error) {
	if err := m.sameCurrency(o); err != nil {
		return Money{}, err
	}
	sum, err := toInt64(new(big.Int).Add(big.NewInt(m.amount), big.NewInt(o.amount)))
	if err != nil {
		return Money{}, err
	}
	return Money{amount: sum, currency: m.currency}, nil
}

func (m Money) Sub(o Money) (Money, error) {
	return m.Add(o.Neg())
}

// Mul multiplies by a whole number, e.g. a unit price by a quantity.
func (m Money) Mul(n int64) (Money, error) {
	p, err := toInt64(new(big.Int).Mul(big.NewInt(m.amount), big.NewInt(n)))
	if err != nil {
		return Money{}, err
	}
	return Money{amount: p, currency: m.currency}, nil
}

// MulRate multiplies by a decimal rate, e.g. "0.2" for 20% VAT.
func (m Money) MulRate(r Rate, mode RoundingMode) (Money, error) {
	num := new(big.Int).Mul(big.NewInt(m.amount), r.v.Num())
	amount, err := toInt64(round(num, r.v.Denom(), mode))
	if err != nil {
		return Money{}, err
	}
	return Money{amount: amount, currency: m.currency}, nil
}

// Convert exchanges m into another currency at rate, where rate is the
// number of target units for one unit of m's currency.
func (m Money) Convert(code string, r Rate, mode RoundingMode) (Money, error) {
	to, err := Lookup(code)
	if err != nil {
		return Money{}, err
	}

	num := new(big.Int).Mul(big.NewInt(m.amount), r.v.Num())
	num.Mul(num, pow10(to.Digits))
	den := new(big.Int).Mul(r.v.Denom(), pow10(m.currency.Digits))

	amount, err := toInt64(round(num, den, mode))
	if err != nil {
		return Money{}, err
	}
	return Money{amount: amount, currency: to}, nil
}

// Allocate splits m by ratios without losing a minor unit. Leftover units go
// to the earliest parts with a non-zero ratio.
func (m Money) Allocate(ratios ...int) ([]Money, error) {
	if len(ratios) == 0 {
		return nil, errors.New("money: no ratios given")
	}

	var total int64
	for _, r := range ratios {
		if r < 0 {
			return nil, errors.New("money: ratios must not be negative")
		}
		total += int64(r)
	}
	if total == 0 {
		return nil, errors.New("money: ratios sum to zero")
	}

	parts := make([]Money, len(ratios))
	rest := m.amount
	for i, r := range ratios {
		share := new(big.Int).Mul(big.NewInt(m.amount), big.NewInt(int64(r)))
		share.Quo(share, big.NewInt(total))
		parts[i] = Money{amount: share.Int64(), currency: m.currency}
		rest -= share.Int64()
	}

	step := int64(1)
	if rest < 0 {
		step = -1
	}
	for i := 0; rest != 0; i = (i + 1) % len(parts) {
		if ratios[i] == 0 {
			continue
		}
		parts[i].amount += step
		rest -= step
	}
	return parts, nil
}

// Split divides m into n parts that differ by at most one minor unit.
func (m Money) Split(n int) ([]Money, error) {
	if n <= 0 {
		return nil, errors.New("money: n must be positive")
	}
	ratios := make([]int, n)
	for i := range ratios {
		ratios[i] = 1
	}
	return m.Allocate(ratios...)
}

// Cmp returns -1, 0 or +1. Comparing different currencies is an error.
func (m Money) Cmp(o Money) (int, error) {
	if err := m.sameCurrency(o); err != nil {
		return 0, err
	}
	switch {
	case m.amount < o.amount:
		return -1, nil
	case m.amount > o.amount:
		return 1, nil
	}
	return 0, nil
}

func (m Money) Equal(o Money) bool {
	return m.currency == o.currency && m.amount == o.amount
}

// Decimal returns the amount in major units, e.g. "12.50".
func (m Money) Decimal() string {
	a := m.amount
	neg := a < 0
	if neg {
		a = -a
	}

	s := strconv.FormatInt(a, 10)
	if d := m.currency.Digits; d > 0 {
		if len(s) <= d {
			s = strings.Repeat("0", d-len(s)+1) + s
		}
		s = s[:len(s)-d] + "." + s[len(s)-d:]
	}
	if neg {
		s = "-" + s
	}
	return s
}

// String returns the amount and code, e.g. "12.50 GBP".
func (m Money) String() string {
	return m.Decimal() + " " + m.currency.Code
}

func (m Money) sameCurrency(o Money) error {
	if m.currency.Code != o.currency.Code {
		return fmt.Errorf("%w: %s and %s", ErrCurrencyMismatch, m.currency.Code, o.currency.Code)
	}
	return nil
}

func toInt64(n *big.Int) (int64, error) {
	if !n.IsInt64() || n.Int64() == math.MinInt64 {
		return 0, ErrOverflow
	}
	return n.Int64(), nil
}

func pow10(n int) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil)
}

func isDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
