package currency

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type rate struct {
	Val  float64   `json:"val"`
	Date time.Time `json:"date"`
}

type rateNBP struct {
	Mid  float64 `json:"mid"`
	Date string  `json:"effectiveDate"`
}
type ratesSeriesNBP struct {
	Code  string    `json:"code"`
	Rates []rateNBP `json:"rates"`
}
type ratesSeries struct {
	BaseCur string `json:"base_cur"`
	Cur     string `json:"cur"`
	Rates   []rate `json:"rates"`
}

func GetRatesSeries(baseCur, cur string, startDate, endDate time.Time) (ratesSeries, error) {
	startDateS := getDate(startDate)
	endDateS := getDate(endDate)
	if baseCur == "PLN" {
		rsNBP, err := callNBP(cur, startDateS, endDateS)
		rr := make([]rate, len(rsNBP.Rates))
		if err != nil {
			return ratesSeries{}, fmt.Errorf("Failed on call to NBP. %v.", err)
		}

		for i, rs := range rsNBP.Rates {
			date, err := time.Parse(time.RFC3339, rs.Date+"T00:00:00Z")
			if err != nil {
				return ratesSeries{}, fmt.Errorf("Failed on parsing date for PLN. %v", err)
			}
			rr[i] = rate{rs.Mid, date}
		}
		return ratesSeries{baseCur, cur, rr}, nil
	}
	rsBaseNBP, err := callNBP(baseCur, startDateS, endDateS)
	if err != nil {
		return ratesSeries{}, fmt.Errorf("Failed on base call to NBP. %v.", err)
	}
	rsCurNBP, err := callNBP(cur, startDateS, endDateS)
	if err != nil {
		return ratesSeries{}, fmt.Errorf("Failed on cur call to NBP. %v.", err)
	}
	rr := make([]rate, len(rsBaseNBP.Rates))
	for i := range rsBaseNBP.Rates {
		date, err := time.Parse(time.RFC3339, rsBaseNBP.Rates[i].Date+"T00:00:00Z")
		if err != nil {
			return ratesSeries{}, fmt.Errorf("Failed on parsing date for not PLN currencies. %v.", err)
		}
		rr[i] = rate{float64(int(10000*rsCurNBP.Rates[i].Mid/rsBaseNBP.Rates[i].Mid)) / 10000, date}
	}
	return ratesSeries{baseCur, cur, rr}, nil

}

func callNBP(cur, startDate, endDate string) (ratesSeriesNBP, error) {
	client := clientWithTimeout(true)
	url := fmt.Sprintf("https://api.nbp.pl/api/exchangerates/rates/A/%v/%v/%v?format=json", cur, startDate, endDate)
	resp, err := client.Get(url)
	defer resp.Body.Close()
	if err != nil {
		return ratesSeriesNBP{}, fmt.Errorf("Failed on GET on NBP api. %v.", err)
	}
	log.Debug("%v succesfully called", url)
	decoder := json.NewDecoder(resp.Body)
	var rsNBP ratesSeriesNBP
	err = decoder.Decode(&rsNBP)
	if err != nil {
		return ratesSeriesNBP{}, fmt.Errorf("Failed on decoding. %v.", err)
	}
	return rsNBP, nil
}

func getDate(date time.Time) string {
	return date.String()[:10]
}
func clientWithTimeout(tlsSecure bool) (client http.Client) {
	timeout := 30 * time.Second
	//Default http client does not have timeout
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !tlsSecure},
	}
	return http.Client{Timeout: timeout, Transport: tr}
}
