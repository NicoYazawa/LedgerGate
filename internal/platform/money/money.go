package money

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"
)

const MicroPerUSD MicroUSD = 1_000_000

const (
	Max MicroUSD = math.MaxInt64
	Min MicroUSD = math.MinInt64
)

var ErrOverflow = errors.New("micro-USD overflow")

type MicroUSD int64

func Parse(value string) (MicroUSD, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("parse micro-USD: empty amount")
	}

	negative := false
	switch value[0] {
	case '-':
		negative = true
		value = value[1:]
	case '+':
		value = value[1:]
	}
	if value == "" || strings.Count(value, ".") > 1 {
		return 0, fmt.Errorf("parse micro-USD %q: invalid decimal", value)
	}

	parts := strings.SplitN(value, ".", 2)
	integerPart := parts[0]
	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
	}
	if integerPart == "" && fraction == "" {
		return 0, fmt.Errorf("parse micro-USD %q: invalid decimal", value)
	}
	for _, part := range []string{integerPart, fraction} {
		for _, character := range part {
			if character < '0' || character > '9' {
				return 0, fmt.Errorf("parse micro-USD %q: invalid digit", value)
			}
		}
	}

	if integerPart == "" {
		integerPart = "0"
	}
	paddedFraction := fraction
	if len(paddedFraction) < 6 {
		paddedFraction += strings.Repeat("0", 6-len(paddedFraction))
	}
	fractionMicros := paddedFraction[:6]

	amount := new(big.Int)
	amount.SetString(integerPart+fractionMicros, 10)
	if len(fraction) > 6 && fraction[6] >= '5' {
		amount.Add(amount, big.NewInt(1))
	}
	if negative {
		amount.Neg(amount)
	}
	if !amount.IsInt64() {
		return 0, fmt.Errorf("parse micro-USD %q: %w", value, ErrOverflow)
	}
	return MicroUSD(amount.Int64()), nil
}

func (amount MicroUSD) String() string {
	integer := new(big.Int).SetInt64(int64(amount))
	sign := ""
	if integer.Sign() < 0 {
		sign = "-"
		integer.Abs(integer)
	}
	divisor := big.NewInt(int64(MicroPerUSD))
	whole := new(big.Int)
	fraction := new(big.Int)
	whole.QuoRem(integer, divisor, fraction)
	return fmt.Sprintf("%s%s.%06s", sign, whole.String(), fraction.String())
}

func (amount MicroUSD) Add(other MicroUSD) (MicroUSD, error) {
	if (other > 0 && amount > Max-other) || (other < 0 && amount < Min-other) {
		return 0, ErrOverflow
	}
	return amount + other, nil
}

func (amount MicroUSD) Sub(other MicroUSD) (MicroUSD, error) {
	if (other > 0 && amount < Min+other) || (other < 0 && amount > Max+other) {
		return 0, ErrOverflow
	}
	return amount - other, nil
}

func (amount MicroUSD) Mul(multiplier int64) (MicroUSD, error) {
	product := new(big.Int).Mul(big.NewInt(int64(amount)), big.NewInt(multiplier))
	if !product.IsInt64() {
		return 0, ErrOverflow
	}
	return MicroUSD(product.Int64()), nil
}
