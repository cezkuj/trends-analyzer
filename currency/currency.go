package currency

import (
	"time"
        "fmt"
)

type rate struct {
	Val  float32   `json:"val"`
	Date time.Time `json:"date"`
}

type rateNBP struct {
	Mid  float32 `json:"mid"`
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
			return ratesSeries{}, err
		}

		for i, rs := range rsNBP.Rates {
			date, err := time.Parse(time.RFC3339, rs.Date+"T00:00:00Z")
			if err != nil {
				return ratesSeries{}, err
			}
			rr[i] = rate{rs.Mid, date}
		}
		return ratesSeries{baseCur, cur, rr}, nil
	}
	rsBaseNBP, err := callNBP(baseCur, startDateS, endDateS)
	if err != nil {
		return ratesSeries{}, err
	}
	rsCurNBP, err := callNBP(cur, startDateS, endDateS)
	if err != nil {
		return ratesSeries{}, err
	}
	rr := make([]rate, len(rsBaseNBP.Rates))
	for i := range rsBaseNBP.Rates {
		date, err := time.Parse(time.RFC3339, rsBaseNBP.Rates[i].Date+"T00:00:00Z")
		if err != nil {
			return ratesSeries{}, err
		}
		rr[i] = rate{rsBaseNBP.Rates[i].Mid / rsCurNBP.Rates[i].Mid, date}
	}
	return ratesSeries{baseCur, cur, rr}, nil

}

func callNBP(cur string, startDate, endDate time.Time) (ratesSeriesNBP, error) {
        client := clientWithTimeout("true")
        url := fmt.Spritntf("https://api.nbp.pl/api/exchangerates/rates/A/%v/%v/%v?format=json", cur, startDate, endDate)
        resp, err := client.Get(url)
        if err != nil {
             return ratesSeriesNBP{}, nil
        }
        defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var rsNBP ratesSeriesNBP
	err = decoder.Decode(&rsNBP)
	if err != nil {
		return nil, err
	}
	return rsNBP, nil
}

func getDate(date time.Time) {
   year, month, day := date.Date()
   return fmt.Sprintf("%v-%v-%v", year, month, day)
}
func clientWithTimeout(tlsSecure bool) (client http.Client) {
        timeout := 30 * time.Second
        //Default http client does not have timeout
        tr := &http.Transport{
                TLSClientConfig: &tls.Config{InsecureSkipVerify: !tlsSecure},
        }
        return http.Client{Timeout: timeout, Transport: tr}

}

