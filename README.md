# money

Currency-safe money for Go. Amounts are stored as integer minor units, rounding is always explicit, and allocation never loses a penny.

```sh
go get github.com/hostfog/money
```

## Usage

```go
price := money.MustParse("19.99", "GBP")
total, _ := price.Mul(3)                                    // 59.97 GBP
vat, _ := total.MulRate(money.MustRate("0.2"), money.HalfUp) // 11.99 GBP

inr, _ := total.Convert("INR", money.MustRate("110.4521"), money.Down)

parts, _ := money.MustParse("100.00", "GBP").Split(3)       // 33.34, 33.33, 33.33
```

## Design

- **No floats.** Amounts are `int64` minor units; rates are exact `big.Rat`.
- **No silent rounding.** `Parse("1.234", "GBP")` is an error. Multiplying by a rate or converting requires a `RoundingMode`.
- **Currencies don't mix.** Adding GBP to USD returns `ErrCurrencyMismatch`.
- **Overflow is an error**, never a wrap-around.
- **Correct minor units.** JPY has 0 digits, KWD has 3. Add others with `Register`.
- **JSON-safe.** Encodes as `{"amount":"12.50","currency":"GBP"}` so JavaScript never rounds through a float.

## Rounding modes

| Mode       | 0.025 | -0.025 |
|------------|-------|--------|
| `HalfEven` | 0.02  | -0.02  |
| `HalfUp`   | 0.03  | -0.03  |
| `Down`     | 0.02  | -0.02  |
| `Up`       | 0.03  | -0.03  |
| `Floor`    | 0.02  | -0.03  |
| `Ceil`     | 0.03  | -0.02  |

## License

MIT
# money
