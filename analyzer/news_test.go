package analyzer

import (
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
)

//TODO: add mocks (wiremock?)
func TestGetNews(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	newsApiKey := os.Getenv("NEWSAPIKEY")
	if newsApiKey == "" {
		t.Fatal("NewsApiKey not set")
	}
	_, err := getNews("Trump", "any", "any", newsApiKey)
	if err != nil {
		t.Fatal(err)
	}
}
