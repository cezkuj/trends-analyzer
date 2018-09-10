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
	c := twitterClient{"http://localhost:8080", twitterAPIKey, clientWithTimeout(false)}
	tweets, err := c.getTweets("trump", "us", "any")
	log.Debug(tweets)
	if err != nil {
		t.Fatal(err)
	}
}
