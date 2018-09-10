package analyzer

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestGetTweets(t *testing.T) {
	c := apiClient{TwitterAPIUrl, "", mockClient{"examples/twitter.json"}}
	tweets, err := c.getTweets("trump", "us", "any")
	log.Info(tweets)
	if err != nil {
		t.Fatal(err)
	}
}
