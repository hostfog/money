package money

import (
	"encoding/json"
	"errors"
	"math"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		in, code string
		want     int64
		err      bool
	}{
		{"12.50", "GBP", 1250, false},
		{"12.5", "GBP", 1250, false},
		{"-0.01", "GBP", -1, false},
		{"+3", "USD", 300, false},
		{".5", "EUR", 50, false},
		{"1000", "JPY", 1000, false},
		{"1.234", "KWD", 1234, false},
		{"1.234", "GBP", 0, true},
		{"1.5", "JPY", 0, true},
		{"abc", "GBP", 0, true},
		{"1,000", "GBP", 0, true},
		{"", "GBP", 0, true},
		{".", "GBP", 0, true},
		{"1", "XXX", 0, true},
		{"99999999999999999999", "GBP", 0, true},
	}
	for _, tt := range tests {
		m, err := Parse(tt.in, tt.code)
		if (err != nil) != tt.err {
			t.Errorf("Parse(%q, %s) err = %v, want err %v", tt.in, tt.code, err, tt.err)
			continue
		}
		if err == nil && m.Amount() != tt.want {
			t.Errorf("Parse(%q, %s) = %d, want %d", tt.in, tt.code, m.Amount(), tt.want)
		}
	}
}

func TestDecimal(t *testing.T) {
	tests := []struct {
		minor int64
		code  string
		want  string
	}{
		{1250, "GBP", "12.50"},
		{5, "GBP", "0.05"},
		{-5, "GBP", "-0.05"},
		{0, "GBP", "0.00"},
		{1000, "JPY", "1000"},
		{1, "KWD", "0.001"},
	}
	for _, tt := range tests {
		m, _ := New(tt.minor, tt.code)
		if got := m.Decimal(); got != tt.want {
			t.Errorf("New(%d, %s).Decimal() = %q, want %q", tt.minor, tt.code, got, tt.want)
		}
	}
}

func TestAddSub(t *testing.T) {
	a := MustParse("10.00", "GBP")
	b := MustParse("2.55", "GBP")

	sum, err := a.Add(b)
	if err != nil || sum.String() != "12.55 GBP" {
		t.Fatalf("Add = %v, %v", sum, err)
	}
	diff, err := b.Sub(a)
	if err != nil || diff.String() != "-7.45 GBP" {
		t.Fatalf("Sub = %v, %v", diff, err)
	}

	if _, err := a.Add(MustParse("1", "USD")); !errors.Is(err, ErrCurrencyMismatch) {
		t.Fatalf("expected currency mismatch, got %v", err)
	}
}

func TestOverflow(t *testing.T) {
	big, _ := New(math.MaxInt64, "GBP")
	one, _ := New(1, "GBP")

	if _, err := big.Add(one); !errors.Is(err, ErrOverflow) {
		t.Errorf("Add overflow: got %v", err)
	}
	if _, err := big.Mul(2); !errors.Is(err, ErrOverflow) {
		t.Errorf("Mul overflow: got %v", err)
	}
	if _, err := New(math.MinInt64, "GBP"); !errors.Is(err, ErrOverflow) {
		t.Errorf("New(MinInt64): got %v", err)
	}
}

func TestRounding(t *testing.T) {
	tests := []struct {
		amount string
		rate   string
		mode   RoundingMode
		want   string
	}{
		{"0.05", "0.5", HalfEven, "0.02"},
		{"0.07", "0.5", HalfEven, "0.04"},
		{"0.05", "0.5", HalfUp, "0.03"},
		{"-0.05", "0.5", HalfUp, "-0.03"},
		{"-0.05", "0.5", HalfEven, "-0.02"},
		{"0.05", "0.5", Down, "0.02"},
		{"0.05", "0.5", Up, "0.03"},
		{"-0.05", "0.5", Floor, "-0.03"},
		{"-0.05", "0.5", Ceil, "-0.02"},
		{"10.00", "1/3", HalfEven, "3.33"},
		{"19.99", "0.2", HalfUp, "4.00"},
	}
	for _, tt := range tests {
		got, err := MustParse(tt.amount, "GBP").MulRate(MustRate(tt.rate), tt.mode)
		if err != nil {
			t.Fatal(err)
		}
		if got.Decimal() != tt.want {
			t.Errorf("%s × %s (mode %d) = %s, want %s", tt.amount, tt.rate, tt.mode, got.Decimal(), tt.want)
		}
	}
}

func TestConvert(t *testing.T) {
	tests := []struct {
		amount, from, to, rate, want string
	}{
		{"100.00", "GBP", "INR", "110.4521", "11045.21 INR"},
		{"100.00", "GBP", "JPY", "191.237", "19124 JPY"},
		{"1000", "JPY", "GBP", "0.00523", "5.23 GBP"},
		{"10.00", "USD", "KWD", "0.3071", "3.071 KWD"},
	}
	for _, tt := range tests {
		got, err := MustParse(tt.amount, tt.from).Convert(tt.to, MustRate(tt.rate), HalfEven)
		if err != nil {
			t.Fatal(err)
		}
		if got.String() != tt.want {
			t.Errorf("Convert %s %s → %s = %s, want %s", tt.amount, tt.from, tt.to, got, tt.want)
		}
	}
}

func TestAllocate(t *testing.T) {
	tests := []struct {
		amount string
		ratios []int
		want   []string
	}{
		{"0.05", []int{70, 30}, []string{"0.04", "0.01"}},
		{"100.00", []int{1, 1, 1}, []string{"33.34", "33.33", "33.33"}},
		{"-100.00", []int{1, 1, 1}, []string{"-33.34", "-33.33", "-33.33"}},
		{"0.03", []int{0, 1, 1}, []string{"0.00", "0.02", "0.01"}},
	}
	for _, tt := range tests {
		m := MustParse(tt.amount, "GBP")
		parts, err := m.Allocate(tt.ratios...)
		if err != nil {
			t.Fatal(err)
		}
		var total int64
		for i, p := range parts {
			total += p.Amount()
			if p.Decimal() != tt.want[i] {
				t.Errorf("Allocate(%s, %v)[%d] = %s, want %s", tt.amount, tt.ratios, i, p.Decimal(), tt.want[i])
			}
		}
		if total != m.Amount() {
			t.Errorf("Allocate(%s, %v) lost money: %d != %d", tt.amount, tt.ratios, total, m.Amount())
		}
	}

	if _, err := MustParse("1", "GBP").Allocate(0, 0); err == nil {
		t.Error("expected error for zero ratios")
	}
}

func TestJSON(t *testing.T) {
	m := MustParse("-12.05", "GBP")
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `{"amount":"-12.05","currency":"GBP"}` {
		t.Fatalf("Marshal = %s", b)
	}

	var got Money
	if err := json.Unmarshal(b, &got); err != nil || !got.Equal(m) {
		t.Fatalf("Unmarshal = %v, %v", got, err)
	}
}

func TestRegister(t *testing.T) {
	if err := Register("XTS", 4); err != nil {
		t.Fatal(err)
	}
	if m := MustParse("1.2345", "xts"); m.Amount() != 12345 {
		t.Errorf("got %d", m.Amount())
	}
	for _, bad := range []string{"gb", "gbp", "G1P"} {
		if Register(bad, 2) == nil {
			t.Errorf("Register(%q) should fail", bad)
		}
	}
}
