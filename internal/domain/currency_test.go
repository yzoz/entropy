package domain

import "testing"

func TestParseAmount(t *testing.T) {
	cases := []struct {
		cur  Currency
		in   string
		want int64
		bad  bool
	}{
		{USD, "12.50", 1250, false},
		{USD, "0.01", 1, false},
		{USD, "-3.00", -300, false},
		{USD, "7", 700, false},
		{USD, "1.234", 0, true}, // too many decimals for a 2-scale currency
		{GLD, "1.2345", 12345, false},
		{OIL, "1.234", 1234, false},
		{USD, "", 0, true},
		{USD, "abc", 0, true},
	}
	for _, c := range cases {
		got, err := ParseAmount(c.cur, c.in)
		if c.bad {
			if err == nil {
				t.Errorf("ParseAmount(%s, %q) expected error, got %d", c.cur, c.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseAmount(%s, %q) unexpected error: %v", c.cur, c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("ParseAmount(%s, %q) = %d, want %d", c.cur, c.in, got, c.want)
		}
	}
}

func TestFormatAmountRoundTrip(t *testing.T) {
	for _, cur := range []Currency{USD, EUR, GLD, SLV, OIL} {
		for _, v := range []int64{0, 1, 99, 100, 12345, -5000} {
			s := FormatAmount(cur, v)
			back, err := ParseAmount(cur, s)
			if err != nil {
				t.Fatalf("round trip %s %d: %v", cur, v, err)
			}
			if back != v {
				t.Errorf("round trip %s %d -> %q -> %d", cur, v, s, back)
			}
		}
	}
}

func TestParseCurrency(t *testing.T) {
	if _, err := ParseCurrency("usd"); err != nil {
		t.Errorf("lowercase currency should be accepted: %v", err)
	}
	if _, err := ParseCurrency("XYZ"); err == nil {
		t.Error("unknown currency should be rejected")
	}
}
