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
	c := twitterClient{TwitterAPIUrl, twitterAPIKey, clientWithTimeout(false)}
	tweets, err := c.getTweets("Trump", "any", "any")
	log.Debug(tweets)
	if err != nil {
		t.Fatal(err)
	}
}
