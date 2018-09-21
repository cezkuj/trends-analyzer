package stock

import (
	"fmt"
	"time"

	av "github.com/cmckee-dev/go-alpha-vantage"
	log "github.com/sirupsen/logrus"
)

type TimeSeriesValue struct {
	Time  time.Time `json:"time"`
	Price float64   `json:"price"`
}

func Series(key, symbol string, startDate, endDate time.Time) ([]TimeSeriesValue, error) {
	client := av.NewClient(key)
	stocksSeries, err := client.StockTimeSeries(av.TimeSeriesDaily, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed on call to StockTimeSeries, %v", err)
	}
	filteredStocks := []TimeSeriesValue{}
	for _, s := range stocksSeries {
		if s.Time.After(endDate) {
			return filteredStocks, nil
		}
		if s.Time.Before(startDate) {
			continue
		}
		filteredStocks = append(filteredStocks, TimeSeriesValue{s.Time, s.Open})
	}
	log.Debug(filteredStocks)
	return filteredStocks, nil
}
