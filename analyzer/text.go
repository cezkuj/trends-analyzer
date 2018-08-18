package analyzer

import (
	"crypto/tls"
	"net/http"
	"time"
)

type text interface {
	Text() string
}

func clientWithTimeout(tlsSecure bool) (client http.Client) {
	timeout := 30 * time.Second
	//Default http client does not have timeout
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !tlsSecure},
	}
	return http.Client{Timeout: timeout, Transport: tr}

}

func analyzeText(text text, provider string) {

}
