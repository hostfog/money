package money_test

import (
	"fmt"

	"github.com/hostfog/money"
)

func Example() {
	price := money.MustParse("19.99", "GBP")
	total, _ := price.Mul(3)
	vat, _ := total.MulRate(money.MustRate("0.2"), money.HalfUp)

	fmt.Println(total, vat)
	// Output: 59.97 GBP 11.99 GBP
}

func ExampleMoney_Convert() {
	sent := money.MustParse("250.00", "GBP")
	received, _ := sent.Convert("INR", money.MustRate("110.4521"), money.Down)

	fmt.Println(received)
	// Output: 27613.02 INR
}

func ExampleMoney_Allocate() {
	fee := money.MustParse("0.05", "GBP")
	parts, _ := fee.Allocate(70, 30)

	fmt.Println(parts[0], parts[1])
	// Output: 0.04 GBP 0.01 GBP
}
