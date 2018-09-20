package stock

import (
	"fmt"
	"time"

	av "github.com/cmckee-dev/go-alpha-vantage"
)

type TimeSeriesValue struct {
	Time  time.Time
	Price float64
}

func Stocks(key, symbol string, startDate, endDate time.Time) ([]TimeSeriesValue, error) {
	client := av.NewClient(key)
	stocksSeries, err := client.StockTimeSeries(av.TimeSeriesDaily, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed on call to StockTimeSeries, %v", err)
	}
	filteredStocks := []TimeSeriesValue{}
	for _, s := range stocksSeries {
		if s.Time.After(endDate) {
			continue
		}
		if s.Time.Before(startDate) {
			return filteredStocks, nil
		}
		filteredStocks = append(filteredStocks, TimeSeriesValue{s.Time, s.Open})
	}
	return filteredStocks, nil
}
