package analyzer

import (
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
)

//TODO: add mocks (wiremock?)
func TestGetNews(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	newsAPIKey := os.Getenv("NEWSAPIKEY")
	if newsAPIKey == "" {
		t.Fatal("NewsAPIKey not set")
	}
	_, err := getNews("Trump", "any", "any", newsAPIKey)
	if err != nil {
		t.Fatal(err)
	}
}
