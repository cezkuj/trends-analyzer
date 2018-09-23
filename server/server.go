package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/cezkuj/trends-analyzer/analyzer"
	"github.com/cezkuj/trends-analyzer/crypto"
	"github.com/cezkuj/trends-analyzer/currency"
	"github.com/cezkuj/trends-analyzer/db"
	"github.com/cezkuj/trends-analyzer/stock"
)

type DbCfg struct {
	user string
	pass string
	host string
	port int
	name string
}

func NewDbCfg(user, pass, host string, port int, name string) DbCfg {
	return DbCfg{user, pass, host, port, name}
}

func StartServer(dbCfg DbCfg, twitterAPIKey, newsAPIKey, stocksAPIKey, salt string, dispatchInterval int, readOnly bool) {
	database, err := db.InitDb(fmt.Sprintf("%v:%v@tcp(%v:%v)/%v", dbCfg.user, dbCfg.pass, dbCfg.host, dbCfg.port, dbCfg.name))
	if err != nil {
		log.Fatal(fmt.Errorf("Failed on InitDb in StartServer, %v", err))
	}
	env := db.NewEnv(database, twitterAPIKey, newsAPIKey, stocksAPIKey, salt)
	go analyzer.StartDispatching(env, dispatchInterval)
	startHttpServer(env, readOnly)
}

type analyzeParams struct {
	keyword         string
	keywordProvider string
	date            string
	country         string
	textProvider    string
}

func analyze(env db.Env) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var dat map[string]string
		err := decoder.Decode(&dat)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Error(fmt.Errorf("Failed on decoding in analyze, %v", err))
			return
		}
		aP, err := parseBody(dat)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Error(fmt.Errorf("Failed on parsing analyze parameters, %v", err))
			return
		}
		k := db.NewKeyword(aP.keyword, aP.keywordProvider, "")
		err = env.CreateKeywordIfNotPresent(k)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Error(fmt.Errorf("Call to CreateKeywordIfNotPresent in analyze, %v", err))
			return
		}
		go analyzer.Analyze(env, k.Name, aP.textProvider, aP.country, aP.date)
	}
}

func parseBody(dat map[string]string) (analyzeParams, error) {
	keyword, present := dat["keyword"]
	if !present {
		return analyzeParams{}, errors.New("Keyword not present in analyze")
	}
	date, present := dat["date"]
	if !present {
		date = "any"
	} else if date != "today" {
		return analyzeParams{}, fmt.Errorf("Date %v not supported", date)
	}
	country, present := dat["country"]
	if !present {
		country = "any"
	} else if country != "pl" && country != "gb" && country != "us" && country != "de" && country != "fr" {
		return analyzeParams{}, fmt.Errorf("Country %v not supported", country)
	}
	keywordProvider, present := dat["keywordProvider"]
	if !present {
		keywordProvider = "unknown"
	}
	textProvider, present := dat["textProvider"]
	if !present {
		textProvider = "both"
	} else if textProvider != "twitter" && textProvider != "news" {
		return analyzeParams{}, errors.New("provider not recognized")
	}
	return analyzeParams{
		keyword:         keyword,
		date:            date,
		country:         country,
		keywordProvider: keywordProvider,
		textProvider:    textProvider,
	}, nil
}

func status(env db.Env) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "showing all statuses.")
	}

}

func keywords(env db.Env) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		keywords, err := env.GetKeywords()
		if err != nil {
			log.Error(fmt.Errorf("Call to GetKeywords failed in keywords, %v", err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		keywordsJSON, err := json.Marshal(keywords)
		if err != nil {
			log.Error(fmt.Errorf("Failed on marshalling %v in keyword, %v", keywords, err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Write(keywordsJSON)
	}

}

func analyzes(env db.Env) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		//Declaring variables beforehand, to bypass scoping problems with if - to refactor later on
		vars := mux.Vars(r)
		keyword := vars["keyword"]
		values := r.URL.Query()
		after, err := parseTime(values.Get("after"), time.Time{})
		if err != nil {
			log.Error(fmt.Errorf("Failed on call to parseTime in analyzes, %v", err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		before, err := parseTime(values.Get("before"), time.Now())
		if err != nil {
			log.Error(fmt.Errorf("Failed on call to parseTime in analyzes, %v", err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		country := values.Get("country")
		if country == "" {
			country = "any"
		}
		keywordPresent, err := env.KeywordIsPresent(keyword)
		if err != nil {
			log.Error(fmt.Errorf("Call to KeywordIsPresent failed in analyzes, %v", err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if !keywordPresent {
			log.Error(fmt.Sprintf("%v is not present", keyword))
			w.WriteHeader(http.StatusNotFound)
			return
		}
		analyzes, err := env.GetAnalyzes(keyword, after, before, country)
		if err != nil {
			log.Error(fmt.Errorf("Call to GetAnalyzes failed in analyzes, %v", err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		analyzesJSON, err := json.Marshal(analyzes)
		if err != nil {
			log.Error(fmt.Errorf("Failed on marshalling %v in analyzes, %v", analyzes, err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Write(analyzesJSON)
	}
}

func parseTime(timeStr string, defaultTime time.Time) (time.Time, error) {
	if timeStr == "" {
		return defaultTime, nil
	}
	parsed, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("Failed on parsing %v in analyzes, %v", timeStr, err)
	}
	return parsed, nil
}

func countries(env db.Env) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		keyword := vars["keyword"]
		analyzes, err := env.GetAnalyzes(keyword, time.Time{}, time.Now(), "any")
		if err != nil {
			log.Error(fmt.Errorf("Call to GetAnalyzes failed in countries, %v", err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		countriesSet := map[string]bool{"any": true}
		for _, a := range analyzes {
			countriesSet[a.Country] = true
		}
		countries := make([]string, len(countriesSet))
		i := 0
		for c := range countriesSet {
			countries[i] = c
			i++
		}
		countriesJSON, err := json.Marshal(countries)
		if err != nil {
			log.Error(fmt.Errorf("Failed on marshaling %v in countries, %v", countries, err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Write(countriesJSON)
	}

}

func stocks(env db.Env) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		symbol := vars["symbol"]
		startDate, endDate, err := parseTimestamps(r.URL.Query())
		if err != nil {
			log.Error(fmt.Errorf("Failed on call to parseTimestamps, %v", err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		stocksSeries, err := stock.Series(env.StocksAPIKey, symbol, startDate, endDate)
		if err != nil {
			log.Error(fmt.Errorf("Call to stocks Series failed in stocks, %v", err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		stocksSeriesJSON, err := json.Marshal(stocksSeries)
		if err != nil {
			log.Error(fmt.Errorf("Failed on marshalling %v in stocks, %v", stocksSeries, err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Write(stocksSeriesJSON)
	}
}

func cryptocurrencies(env db.Env) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		fromCurrency := vars["fromCurrency"]
		toCurrency := vars["toCurrency"]
		startDate, endDate, err := parseTimestamps(r.URL.Query())
		if err != nil {
			log.Error(fmt.Errorf("Failed on call to parseTimestamps, %v", err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		cryptoSeries, err := crypto.Series(env.StocksAPIKey, fromCurrency, toCurrency, startDate, endDate)
		if err != nil {
			log.Error(fmt.Errorf("Call to crypto Series failed in crypto, %v", err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		cryptoSeriesJSON, err := json.Marshal(cryptoSeries)
		if err != nil {
			log.Error(fmt.Errorf("Failed on marshalling %v in crypto, %v", cryptoSeries, err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Write(cryptoSeriesJSON)
	}
}

func rates(env db.Env) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		baseCur := vars["baseCur"]
		cur := vars["cur"]
		startDate, endDate, err := parseTimestamps(r.URL.Query())
		if err != nil {
			log.Error(fmt.Errorf("Failed on call to parseTimestamps, %v", err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		ratesSeries, err := currency.GetRatesSeries(baseCur, cur, startDate, endDate)
		if err != nil {
			log.Error(fmt.Errorf("Call to GetRatesSeries failed in rates, %v", err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		ratesSeriesJSON, err := json.Marshal(ratesSeries)
		if err != nil {
			log.Error(fmt.Errorf("Failed on marshalling %v in rates, %v", ratesSeries, err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Write(ratesSeriesJSON)
	}
}

func parseTimestamp(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Now(), nil
	}
	parsed, err := strconv.ParseInt(timeStr, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("Failed on parsing timestamp, %v", err)
	}
	return time.Unix(parsed/1000, 0), nil
}

func parseTimestamps(values url.Values) (time.Time, time.Time, error) {
	startDate, err := parseTimestamp(values.Get("startDate"))
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("Failed on call to parseTimestamp, %v", err)
	}
	endDate, err := parseTimestamp(values.Get("endDate"))
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("Failed on call to parseTimestamp, %v", err)
	}
	return startDate, endDate, nil

}

func startHttpServer(env db.Env, readOnly bool) {
	serveMux := createServeMux(env, readOnly)
	srv := &http.Server{
		Addr:         ":8000",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      serveMux,
	}
	log.Println(srv.ListenAndServe())
}

func createServeMux(env db.Env, readOnly bool) *http.ServeMux {
	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api").Subrouter()
	if !readOnly {
		apiRouter.HandleFunc("/analyze", analyze(env)).Methods("POST")
	}
	apiRouter.HandleFunc("/status", status(env)).Methods("GET")
	apiRouter.HandleFunc("/keywords", keywords(env)).Methods("GET")
	apiRouter.HandleFunc("/analyzes/{keyword}", analyzes(env)).Methods("GET")
	apiRouter.HandleFunc("/countries/{keyword}", countries(env)).Methods("GET")
	apiRouter.HandleFunc("/rates/{baseCur}/{cur}", rates(env)).Methods("GET")
	apiRouter.HandleFunc("/stocks/{symbol}", stocks(env)).Methods("GET")
	apiRouter.HandleFunc("/crypto/{fromCurrency}/{toCurrency}", cryptocurrencies(env)).Methods("GET")
	serveMux := &http.ServeMux{}
	serveMux.Handle("/", router)
	return serveMux
}
