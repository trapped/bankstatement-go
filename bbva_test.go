package bankstatement

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseReportDate(t *testing.T) {
	tcs := []struct {
		v   string
		out time.Time
		err bool
	}{
		{"12/04/1987", time.Date(1987, 4, 12, 0, 0, 0, 0, time.UTC), false},
		{"31/01/2022", time.Date(2022, 1, 31, 0, 0, 0, 0, time.UTC), false},
		{"ababe", time.Time{}, true},
		{"12/04", time.Time{}, true},
	}
	tr := new(BBVAPDFTransactionReader)
	for _, tc := range tcs {
		t.Run(fmt.Sprintf("parses %s", tc.v), func(t *testing.T) {
			out, err := tr.parseReportDate(tc.v)
			assert.Equal(t, tc.out, out, "time doesn't match")
			assert.Equal(t, tc.err, err != nil, "error doesn't match")
		})
	}
}

func TestParseTransactionDate(t *testing.T) {
	tcs := []struct {
		reportDate time.Time
		v          string
		out        time.Time
		err        bool
	}{
		{time.Date(1987, 5, 1, 0, 0, 0, 0, time.UTC), "12/04", time.Date(1987, 4, 12, 0, 0, 0, 0, time.UTC), false},
		{time.Date(2022, 2, 1, 0, 0, 0, 0, time.UTC), "31/01", time.Date(2022, 1, 31, 0, 0, 0, 0, time.UTC), false},
		{time.Date(2022, 2, 1, 0, 0, 0, 0, time.UTC), "ababe", time.Time{}, true},
		{time.Date(2022, 2, 1, 0, 0, 0, 0, time.UTC), "12", time.Time{}, true},
	}
	tr := new(BBVAPDFTransactionReader)
	for _, tc := range tcs {
		t.Run(fmt.Sprintf("parses %s (%s)", tc.v, tc.reportDate), func(t *testing.T) {
			out, err := tr.parseTransactionDate(tc.reportDate, tc.v)
			assert.Equal(t, tc.out, out, "time doesn't match")
			assert.Equal(t, tc.err, err != nil, "error doesn't match")
		})
	}
}

func TestParseAmount(t *testing.T) {
	tcs := []struct {
		v   string
		out float64
		err bool
	}{
		{"127,33", 127.33, false},
		{"1.010,99", 1010.99, false},
		{"ababe", 0.0, true},
		{"12-11", 0.0, true},
	}
	tr := new(BBVAPDFTransactionReader)
	for _, tc := range tcs {
		t.Run(fmt.Sprintf("parses %s", tc.v), func(t *testing.T) {
			out, err := tr.parseAmount(tc.v)
			assert.Equal(t, tc.out, out, "amount doesn't match")
			assert.Equal(t, tc.err, err != nil, "error doesn't match")
		})
	}
}
