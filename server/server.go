package server

import (
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
		log.Fatal(err)
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
			log.Error(err)
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

		tagProvider, present := dat["tagProvider"]
		if !present {
			tagProvider = "unknown"
		}
		tag := db.NewTag(keyword, tagProvider, "")
		err = env.CreateTagIfNotPresent(tag)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Error(err)
			return
		}
		tagID, err := env.GetTagID(tag.Name)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Error(err)
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
		go analyzer.Analyze(env, keyword, textProvider, country, date, tagID)
	}
}

func status(env db.Env) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "showing all statuses.")
	}

}

func tags(env db.Env) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		tags, err := env.GetTags()
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		tagsJSON, err := json.Marshal(tags)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Write(tagsJSON)
	}

}

func analyzes(env db.Env) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		//Declaring variables beforehand, to bypass scoping problems with if - to refactor later on
		var after time.Time
		var before time.Time
		var err error
		values := r.URL.Query()
		name := values.Get("name")
		if name == "" {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "name parameter is missing")
			return
		}
		if afterStr := values.Get("after"); afterStr != "" {
			after, err = time.Parse(time.RFC3339, afterStr)
			if err != nil {
				log.Error(err)
				w.WriteHeader(http.StatusBadRequest)
				return

			}

		} else {
			after = time.Time{}
		}
		if beforeStr := values.Get("before"); beforeStr != "" {
			before, err = time.Parse(time.RFC3339, beforeStr)
			if err != nil {
				log.Error(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

		} else {
			before = time.Now()
		}
		analyzes, err := env.GetAnalyzes(name, after, before)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		analyzesJSON, err := json.Marshal(analyzes)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Write(analyzesJSON)
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
	apiRouter.HandleFunc("/tags", tags(env)).Methods("GET")
	apiRouter.HandleFunc("/analyzes", analyzes(env)).Methods("GET")
	serveMux := &http.ServeMux{}
	serveMux.Handle("/", router)
	return serveMux
}
