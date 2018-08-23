package analyzer

import (
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
)

//TODO: add mocks (wiremock?)
func TestGetTweets(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	twitterApiKey := os.Getenv("TWITTERAPIKEY")
	if twitterApiKey == "" {
		t.Fatal("TwitterApiKey not set")
	}
	tweets, err := getTweets("Trump", "any", "any", twitterApiKey)
	log.Debug(tweets)
	if err != nil {
		t.Fatal(err)
	}
}
