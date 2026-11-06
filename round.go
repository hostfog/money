package money

import "math/big"

// RoundingMode controls how a fractional minor unit is resolved.
type RoundingMode int

const (
	HalfEven RoundingMode = iota // banker's rounding
	HalfUp                       // ties away from zero
	Down                         // toward zero
	Up                           // away from zero
	Floor                        // toward negative infinity
	Ceil                         // toward positive infinity
)

var bigOne = big.NewInt(1)

// round returns num/den rounded to an integer. den must be positive.
func round(num, den *big.Int, mode RoundingMode) *big.Int {
	q, r := new(big.Int).QuoRem(num, den, new(big.Int))
	if r.Sign() == 0 {
		return q
	}

	neg := num.Sign() < 0
	var away bool

	switch mode {
	case Up:
		away = true
	case Floor:
		away = neg
	case Ceil:
		away = !neg
	case HalfUp, HalfEven:
		twice := new(big.Int).Abs(r)
		twice.Lsh(twice, 1)
		c := twice.Cmp(den)
		away = c > 0 || (c == 0 && (mode == HalfUp || q.Bit(0) == 1))
	}

	if away {
		if neg {
			q.Sub(q, bigOne)
		} else {
			q.Add(q, bigOne)
		}
	}
	return q
}
