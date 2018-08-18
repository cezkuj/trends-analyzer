package server

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/acme/autocert"
	"io"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/cezkuj/trends-analyzer/analyzer"
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
	db, err := initDb(dbCfg.user + ":" + dbCfg.pass + "@tcp(" + dbCfg.host + ")/" + dbCfg.name)
	if err != nil {
		log.Fatal(err)
	}
	env := Env{db: db, twitterApiKey: twitterApiKey, newsApiKey: newsApiKey}
	if prod {
		startProdServer(env)
	}
	startDevServer(env)
}

func scrap(env Env) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var dat map[string]string
		err := decoder.Decode(&dat)
		if err != nil {
			log.Println(err)
			return
		}
		keyword, present := dat["keyword"]
		if !present {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "Keyword not present")
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
		provider, present := dat["provider"]
		if !present {
			provider = "both"
		}
		switch provider {
		case "twitter":
			go analyzer.AnalyzeTwitter(keyword, country, date, env.twitterApiKey)
		case "news":
			go analyzer.AnalyzeNews(keyword, country, date, env.newsApiKey)
		case "both":
			go analyzer.AnalyzeTwitter(keyword, country, date, env.twitterApiKey)
			go analyzer.AnalyzeNews(keyword, country, date, env.newsApiKey)
		default:
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "Provider not recognized.")
		}
	}
}

func status(env Env) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "showing all statuses.")
	}

}

func startProdServer(env Env) {
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
func startDevServer(env Env) {
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
func createServeMux(env Env) *http.ServeMux {
	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/scrap", scrap(env)).Methods("POST")
	apiRouter.HandleFunc("/status", status(env)).Methods("GET")
	serveMux := &http.ServeMux{}
	serveMux.Handle("/", router)
	return serveMux
}

func parseReaderToJson(reader io.Reader) (map[string]string, error) {
	var dat map[string]string
	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)
	err := json.Unmarshal(buf.Bytes(), &dat)
	return dat, err
}
