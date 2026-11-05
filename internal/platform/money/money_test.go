package money

import "testing"

func TestParseConvertsUSDToMicroUSD(t *testing.T) {
	tests := map[string]MicroUSD{
		"0":                    0,
		"1":                    MicroPerUSD,
		"1.25":                 1_250_000,
		"-1.25":                -1_250_000,
		"+2.000001":            2_000_001,
		".5":                   500_000,
		"1.":                   MicroPerUSD,
		"0.1234564":            123_456,
		"0.1234565":            123_457,
		"-0.1234565":           -123_457,
		"9223372036854.775807": Max,
	}

	for input, want := range tests {
		got, err := Parse(input)
		if err != nil {
			t.Errorf("Parse(%q) error = %v", input, err)
			continue
		}
		if got != want {
			t.Errorf("Parse(%q) = %d, want %d", input, got, want)
		}
	}
}

func TestParseRejectsInvalidAndOverflowingValues(t *testing.T) {
	inputs := []string{
		"",
		" ",
		".?",
		"1.2.3",
		"1e6",
		"abc",
		"9223372036854.775808",
		"-9223372036854.775809",
	}
	for _, input := range inputs {
		if _, err := Parse(input); err == nil {
			t.Errorf("Parse(%q) error = nil, want error", input)
		}
	}
}

func TestStringUsesSixDecimalUSD(t *testing.T) {
	tests := map[MicroUSD]string{
		0:           "0.000000",
		MicroPerUSD: "1.000000",
		1:           "0.000001",
		-1_250_000:  "-1.250000",
		Max:         "9223372036854.775807",
		Min:         "-9223372036854.775808",
	}
	for value, want := range tests {
		if got := value.String(); got != want {
			t.Errorf("%d.String() = %q, want %q", int64(value), got, want)
		}
	}
}

func TestArithmeticChecksOverflow(t *testing.T) {
	if got, err := MicroUSD(2).Add(3); err != nil || got != 5 {
		t.Errorf("Add() = %d, %v; want 5, nil", got, err)
	}
	if _, err := Max.Add(1); err == nil {
		t.Error("Add() overflow error = nil")
	}
	if got, err := MicroUSD(5).Sub(3); err != nil || got != 2 {
		t.Errorf("Sub() = %d, %v; want 2, nil", got, err)
	}
	if _, err := Min.Sub(1); err == nil {
		t.Error("Sub() overflow error = nil")
	}
	if got, err := MicroUSD(3).Mul(-2); err != nil || got != -6 {
		t.Errorf("Mul() = %d, %v; want -6, nil", got, err)
	}
	if _, err := Max.Mul(2); err == nil {
		t.Error("Mul() overflow error = nil")
	}
}
