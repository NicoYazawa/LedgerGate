package money

import (
	"fmt"
	"math"
	"math/big"
	"regexp"
	"strings"
)

const MicroPerUSD MicroUSD = 1_000_000

var decimalPattern = regexp.MustCompile(`^[+-]?(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)$`)

type MicroUSD int64

func Parse(value string) (MicroUSD, error) {
	value = strings.TrimSpace(value)
	if !decimalPattern.MatchString(value) {
		return 0, fmt.Errorf("parse micro-USD %q: invalid decimal", value)
	}
	rational, _ := new(big.Rat).SetString(value)
	rational.Mul(rational, big.NewRat(int64(MicroPerUSD), 1))
	amount, remainder := new(big.Int), new(big.Int)
	amount.QuoRem(rational.Num(), rational.Denom(), remainder)
	if new(big.Int).Lsh(new(big.Int).Abs(remainder), 1).Cmp(rational.Denom()) >= 0 {
		if rational.Sign() < 0 {
			amount.Sub(amount, big.NewInt(1))
		} else {
			amount.Add(amount, big.NewInt(1))
		}
	}
	if !amount.IsInt64() {
		return 0, fmt.Errorf("parse micro-USD %q: overflow", value)
	}
	return MicroUSD(amount.Int64()), nil
}

func (amount MicroUSD) String() string {
	sign := ""
	magnitude := uint64(amount)
	if amount < 0 {
		sign = "-"
		magnitude = uint64(-(amount + 1)) + 1
	}
	return fmt.Sprintf("%s%d.%06d", sign, magnitude/uint64(MicroPerUSD), magnitude%uint64(MicroPerUSD))
}

func (amount MicroUSD) Add(other MicroUSD) (MicroUSD, error) {
	if (other > 0 && amount > MicroUSD(math.MaxInt64)-other) || (other < 0 && amount < MicroUSD(math.MinInt64)-other) {
		return 0, fmt.Errorf("micro-USD overflow")
	}
	return amount + other, nil
}

func (amount MicroUSD) Sub(other MicroUSD) (MicroUSD, error) {
	if (other > 0 && amount < MicroUSD(math.MinInt64)+other) || (other < 0 && amount > MicroUSD(math.MaxInt64)+other) {
		return 0, fmt.Errorf("micro-USD overflow")
	}
	return amount - other, nil
}

func (amount MicroUSD) Mul(multiplier int64) (MicroUSD, error) {
	product := new(big.Int).Mul(big.NewInt(int64(amount)), big.NewInt(multiplier))
	if !product.IsInt64() {
		return 0, fmt.Errorf("micro-USD overflow")
	}
	return MicroUSD(product.Int64()), nil
}
