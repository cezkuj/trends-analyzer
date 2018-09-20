package crypto

import (
	"fmt"
	"time"

	av "github.com/cmckee-dev/go-alpha-vantage"
)

type DigitalCurrencySeriesValue struct {
	Time  time.Time
	Price float64
}

func Crypto(key, fromCurrency, toCurrency string, startDate, endDate time.Time) ([]DigitalCurrencySeriesValue, error) {
	client := av.NewClient(key)
	cryptoSeries, err := client.DigitalCurrency(fromCurrency, toCurrency)
	if err != nil {
		return nil, fmt.Errorf("failed on call to DigitalCurrency, %v", err)
	}
	filteredCrypto := []DigitalCurrencySeriesValue{}
	for _, s := range cryptoSeries {
		if s.Time.After(endDate) {
			continue
		}
		if s.Time.Before(startDate) {
			return filteredCrypto, nil
		}
		filteredCrypto = append(filteredCrypto, DigitalCurrencySeriesValue{s.Time, s.Price})
	}
	return filteredCrypto, nil
}
