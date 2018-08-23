package analyzer

import (
	"crypto/tls"
	"net/http"
	"sync"
	"time"

	language "cloud.google.com/go/language/apiv1"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	languagepb "google.golang.org/genproto/googleapis/cloud/language/v1"

	"github.com/cezkuj/trends-analyzer/db"
)

type text struct {
	id           int
	text         string
	textProvider string
	timestamp    time.Time
}

type analyzedText struct {
	reaction     float32
	textProvider string
	timestamp    time.Time
}

func clientWithTimeout(tlsSecure bool) (client http.Client) {
	timeout := 30 * time.Second
	//Default http client does not have timeout
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !tlsSecure},
	}
	return http.Client{Timeout: timeout, Transport: tr}

}

func Analyze(env db.Env, keyword, textProvider, country, date string, tagID int) {
	tt := []text{}
	if textProvider == "twitter" || textProvider == "both" {
		tweets, err := getTweets(keyword, country, date, env.TwitterApiKey)
		if err != nil {
			log.Error(err)
			return
		}
		tt = append(tt, tweets...)
	}
	if textProvider == "news" || textProvider == "both" {
		nn, err := getNews(keyword, country, date, env.NewsApiKey)
		if err != nil {
			log.Error(err)
			return
		}
		tt = append(tt, nn...)
	}
	c := make(chan analyzedText)
	wg := new(sync.WaitGroup)

	for _, t := range tt {
		wg.Add(1)
		go analyzeText(t, c, wg)
	}
	go func() {
		wg.Wait()
		close(c)

	}()
	var timestampFirst time.Time
	var timestampLast time.Time
	stats := map[string]int{}
	sums := map[string]float32{}
	reactionTweets := float32(0)
	reactionNews := float32(0)
	reactionAvg := float32(0)
	for t := range c {
		stats[t.textProvider] += 1
		sums[t.textProvider] += t.reaction
		log.Debug(timestampFirst, timestampLast, t.timestamp)
		if t.timestamp.After(timestampLast) {
			timestampLast = t.timestamp
		}
		if timestampFirst.Before(time.Date(1000, time.January, 0, 0, 0, 0, 0, time.UTC)) || t.timestamp.Before(timestampFirst) {
			timestampFirst = t.timestamp
		}
	}
	if stats["twitter"] > 0 {
		reactionTweets = sums["twitter"] / float32(stats["twitter"])
		reactionAvg = reactionTweets
	}
	if stats["news"] > 0 {
		reactionNews = sums["news"] / float32(stats["news"])
		reactionAvg = reactionNews
	}
	if stats["twitter"] > 0 && stats["news"] > 0 {
		reactionAvg = (sums["twitter"] + sums["news"]) / float32(stats["twitter"]+stats["news"])
	}
	analyzis := db.NewAnalyzis(tagID, timestampFirst, timestampLast, stats["twitter"], stats["news"], reactionAvg, reactionTweets, reactionNews)
	err := env.CreateAnalyzis(analyzis)
	if err != nil {
		log.Error(err)

	}

}
func analyzeText(t text, c chan analyzedText, wg *sync.WaitGroup) {
	defer wg.Done()
	s, err := analyzeSentiment(t.text)
	if err != nil {
		log.Error(err)
		return
	}
	c <- analyzedText{s, t.textProvider, t.timestamp}

}
func analyzeSentiment(text string) (float32, error) {
	ctx := context.Background()

	// Creates a client.
	client, err := language.NewClient(ctx)
	if err != nil {
		return 0, err
	}
	// Detects the sentiment of the text.
	sentiment, err := client.AnalyzeSentiment(ctx, &languagepb.AnalyzeSentimentRequest{
		Document: &languagepb.Document{
			Source: &languagepb.Document_Content{
				Content: text,
			},
			Type: languagepb.Document_PLAIN_TEXT,
		},
		EncodingType: languagepb.EncodingType_UTF8,
	})
	if err != nil {
		return 0, err
	}
	return sentiment.DocumentSentiment.Score, nil
}
