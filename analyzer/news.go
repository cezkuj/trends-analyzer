package analyzer

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

type news struct {
	id        string
	text      string
	timestamp time.Time
}

func (n news) Text() string {
	return n.text
}

func AnalyzeNews(keyword, country, date, newsApiKey string) {
	news := getNews(keyword, country, date)
	for _, n := range news {
		log.Debug(fmt.Sprintf("Analyzing news %v - %v", n.id, n.timestamp))
		go analyzeText(n, "news")
	}
}

func getNews(keyword, country, date string) []news {
	return nil
}
