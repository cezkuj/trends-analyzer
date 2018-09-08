package analyzer

import (
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
)

//TODO: add mocks (wiremock?)
func TestGetTweets(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	twitterAPIKey := os.Getenv("TWITTERAPIKEY")
	if twitterAPIKey == "" {
		t.Fatal("TwitterAPIKey not set")
	}
	tweets, err := getTweets("Trump", "any", "any", twitterAPIKey)
	log.Debug(tweets)
	if err != nil {
		t.Fatal(err)
	}
}
