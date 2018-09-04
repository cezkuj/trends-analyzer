package currency

import (
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

func TestCallNBP(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	rsNBP, err := callNBP("USD", "2018-01-02", "2018-01-02")
	if err != nil {
		t.Fatal(err)
	}
	expected := ratesSeriesNBP{"USD", []rateNBP{{3.4546, "2018-01-02"}}}
	t.Log(rsNBP)
	if rsNBP.Code != expected.Code || rsNBP.Rates[0] != expected.Rates[0] {
		t.Fatalf("Codes: %v, %v, Rates: %v, %v. Rates series does not match expected ones", rsNBP.Code, expected.Code, rsNBP.Rates[0], expected.Rates[0])
	}
}

func TestGetRatesSeries(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	testCases := []struct {
		baseCur       string
		cur           string
		startDate     time.Time
		endDate       time.Time
		expectedValue float64
	}{
		{"PLN", "USD", time.Date(2018, 01, 02, 0, 0, 0, 0, time.UTC), time.Date(2018, 01, 02, 0, 0, 0, 0, time.UTC), 3.4546},
		{"EUR", "USD", time.Date(2018, 01, 02, 0, 0, 0, 0, time.UTC), time.Date(2018, 01, 02, 0, 0, 0, 0, time.UTC), 0.8284213807822355},
		{"USD", "EUR", time.Date(2018, 01, 02, 0, 0, 0, 0, time.UTC), time.Date(2018, 01, 02, 0, 0, 0, 0, time.UTC), 1.2071151508134081},
	}
	for _, tc := range testCases {
		rs, err := GetRatesSeries(tc.baseCur, tc.cur, tc.startDate, tc.endDate)
		if err != nil {
			t.Fatal(err)
		}
		if rs.Rates[0].Val != tc.expectedValue {
			t.Fatalf("Test case: %v, actual value: %v Rates series does not match expected ones", tc, rs.Rates[0].Val)
		}
	}
}
