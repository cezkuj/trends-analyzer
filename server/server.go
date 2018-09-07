package server

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/acme/autocert"
	"io"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/cezkuj/trends-analyzer/analyzer"
	"github.com/cezkuj/trends-analyzer/currency"
	"github.com/cezkuj/trends-analyzer/db"
)

type DbCfg struct {
	user string
	pass string
	host string
	name string
}

func NewDbCfg(user, pass, host, name string) DbCfg {
	return DbCfg{user, pass, host, name}
}

func StartServer(dbCfg DbCfg, twitterApiKey, newsApiKey string, prod bool) {
	database, err := db.InitDb(dbCfg.user + ":" + dbCfg.pass + "@tcp(" + dbCfg.host + ")/" + dbCfg.name)
	if err != nil {
		log.Fatal(fmt.Errorf("Failed on InitDb in StartServer, %v", err))
	}
	env := db.NewEnv(database, twitterApiKey, newsApiKey)
	if prod {
		startProdServer(env)
	}
	startDevServer(env)
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
		keyword, present := dat["keyword"]
		if !present {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "Keyword not present")
			log.Error("Keyword not present in analyze")
			return
		}
		date, present := dat["date"]
		if !present {
			date = "any"
		} else if date != "today" {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, fmt.Sprintf("Date %v not supported", date))
			return
		}

		country, present := dat["country"]
		if !present {
			country = "any"
		} else if country != "pl" && country != "gb" && country != "us" && country != "de" && country != "fr" {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, fmt.Sprintf("Country %v not supported", country))
			return
		}

		keywordProvider, present := dat["keywordProvider"]
		if !present {
			keywordProvider = "unknown"
		}
		k := db.NewKeyword(keyword, keywordProvider, "")
		err = env.CreateKeywordIfNotPresent(k)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Error(fmt.Errorf("Call to CreateKeywordIfNotPresent in analyze, %v", err))
			return
		}
		textProvider, present := dat["textProvider"]
		if !present {
			textProvider = "both"
		}
		if textProvider != "both" && textProvider != "twitter" && textProvider != "news" {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "Provider not recognized.")
		}
		go analyzer.Analyze(env, k.Name, textProvider, country, date)
	}
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
			log.Error(fmt.Errorf("Call to GetKeywords failed in keywords", err))
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

func dispatcher(env db.Env) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		go analyzer.StartDispatching(env)
	}
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

func rates(env db.Env) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		baseCur := vars["baseCur"]
		cur := vars["cur"]
		values := r.URL.Query()
		startDateS := values.Get("startDate")
		endDateS := values.Get("endDate")
		var startDate time.Time
		var endDate time.Time
		var err error
		if startDateS == "" {
			startDate = time.Now()
		} else {
			startDateI, err := strconv.ParseInt(startDateS, 10, 64)
			if err != nil {
				log.Error(fmt.Errorf("Failed on parsing startDate timestamp, %v", err))
				return
			}
			startDate = time.Unix(startDateI/1000, 0)
		}
		if endDateS == "" {
			endDate = time.Now()
		} else {
			endDateI, err := strconv.ParseInt(endDateS, 10, 64)
			if err != nil {
				log.Error(fmt.Errorf("Failed on parsing endDate timestamp, %v", err))
				return
			}
			endDate = time.Unix(endDateI/1000, 0)

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

func startProdServer(env db.Env) {
	m := &autocert.Manager{
		Cache:      autocert.DirCache(".secret"),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist("gopage.cezkuj.com"),
	}

	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
			tls.X25519,
		},
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
		GetCertificate: m.GetCertificate,
	}
	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		Handler:      m.HTTPHandler(nil),
	}
	go func() { log.Fatal(srv.ListenAndServe()) }()
	serveMux := createServeMux(env)
	srvTLS := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		TLSConfig:    tlsConfig,
		Handler:      serveMux,
	}
	log.Println(srvTLS.ListenAndServeTLS("", ""))

}

func startDevServer(env db.Env) {
	serveMux := createServeMux(env)
	srv := &http.Server{
		Addr:         ":8000",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      serveMux,
	}
	log.Println(srv.ListenAndServe())
}

func createServeMux(env db.Env) *http.ServeMux {
	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/analyze", analyze(env)).Methods("POST")
	apiRouter.HandleFunc("/status", status(env)).Methods("GET")
	apiRouter.HandleFunc("/keywords", keywords(env)).Methods("GET")
	apiRouter.HandleFunc("/analyzes/{keyword}", analyzes(env)).Methods("GET")
	apiRouter.HandleFunc("/countries/{keyword}", countries(env)).Methods("GET")
	apiRouter.HandleFunc("/rates/{baseCur}/{cur}", rates(env)).Methods("GET")
	apiRouter.HandleFunc("/dispatcher", dispatcher(env)).Methods("POST")
	serveMux := &http.ServeMux{}
	serveMux.Handle("/", router)
	return serveMux
}
