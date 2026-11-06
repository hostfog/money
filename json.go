package money

import "encoding/json"

type wire struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

// MarshalJSON encodes as {"amount":"12.50","currency":"GBP"}. The amount is a
// string so JavaScript clients never round it through a float.
func (m Money) MarshalJSON() ([]byte, error) {
	return json.Marshal(wire{Amount: m.Decimal(), Currency: m.currency.Code})
}

func (m *Money) UnmarshalJSON(b []byte) error {
	var w wire
	if err := json.Unmarshal(b, &w); err != nil {
		return err
	}
	v, err := Parse(w.Amount, w.Currency)
	if err != nil {
		return err
	}
	*m = v
	return nil
}
