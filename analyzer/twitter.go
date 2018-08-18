package analyzer

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

type tweet struct {
	id        string
	text      string
	timestamp time.Time
}

func (t tweet) Text() string {
	return t.text
}

func AnalyzeTwitter(keyword, country, date, twitterApiKey string) {
	tweets := getTweets(keyword, country, date)
	for _, t := range tweets {
		log.Debug(fmt.Sprintf("Analyzing tweet %v - %v", t.id, t.timestamp))
		go analyzeText(t, "twitter")
	}
}

func getTweets(keyword, country, date string) []tweet {
	return nil
}
