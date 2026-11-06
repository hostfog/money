package money

import (
	"fmt"
	"math/big"
)

// Rate is an exact decimal multiplier such as an exchange rate or tax rate.
type Rate struct {
	v *big.Rat
}

// ParseRate reads a decimal ("1.1735") or fraction ("1/3"). Rates must be positive.
func ParseRate(s string) (Rate, error) {
	v, ok := new(big.Rat).SetString(s)
	if !ok || v.Sign() <= 0 {
		return Rate{}, fmt.Errorf("money: invalid rate %q", s)
	}
	return Rate{v: v}, nil
}

// MustRate is like ParseRate but panics on error.
func MustRate(s string) Rate {
	r, err := ParseRate(s)
	if err != nil {
		panic(err)
	}
	return r
}

func (r Rate) String() string {
	return r.v.RatString()
}
