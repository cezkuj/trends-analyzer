package analyzer

import (
	"crypto/tls"
	"fmt"
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

func Analyze(env db.Env, keyword, textProvider, country, date string) {
	tt, err := getText(env, keyword, textProvider, country, date)
	if err != nil {
		log.Error(fmt.Errorf("Analyze failed, %v", err))
		return
	}
	c := make(chan analyzedText)
	wg := new(sync.WaitGroup)
	ctx := context.Background()
	client, err := language.NewClient(ctx)
	if err != nil {
		log.Error(fmt.Errorf("Failed to create new language client, %v", err))
		return
	}
	for _, t := range tt {
		wg.Add(1)
		go analyzeText(client, ctx, t, c, wg)
	}
	go func() {
		wg.Wait()
		close(c)
	}()
	count := map[string]int{}
	sums := map[string]float32{}
	for t := range c {
		count[t.textProvider]++
		sums[t.textProvider] += t.reaction
	}
	reactionAvg, reactionTweets, reactionNews := calcReaction(count, sums)
	keywordID, err := env.GetKeywordID(keyword)
	if err != nil {
		log.Error(fmt.Errorf("Failed on call to GetKeywordID for %v in Analyze, %v", keyword, err))
		return
	}
	analyzis := db.NewAnalyzis(keywordID, country, time.Now(), count["twitter"], count["news"], reactionAvg, reactionTweets, reactionNews)
	err = env.CreateAnalyzis(analyzis)
	if err != nil {
		log.Error(fmt.Errorf("Failed on call to CreateAnalyzes for %v in Analyze, %v", analyzis, err))
	}
}

func getText(env db.Env, keyword, textProvider, country, date string) ([]text, error) {
	tt := []text{}
	if textProvider == "twitter" || textProvider == "both" {
		tweets, err := getTweets(keyword, country, date, env.TwitterApiKey)
		if err != nil {
			return nil, fmt.Errorf("Failed on call to getTweets in Analyze, %v", err)
		}
		tt = append(tt, tweets...)
	}
	if textProvider == "news" || textProvider == "both" {
		nn, err := getNews(keyword, country, date, env.NewsApiKey)
		if err != nil {
			return nil, fmt.Errorf("Failed on call to getNews in Analyze, %v", err)
		}
		tt = append(tt, nn...)
	}
	return tt, nil
}

func calcReaction(count map[string]int, sums map[string]float32) (float32, float32, float32) {
	reactionTweets := float32(0)
	reactionNews := float32(0)
	reactionAvg := float32(0)

	if count["twitter"] > 0 {
		reactionTweets = sums["twitter"] / float32(count["twitter"])
		reactionAvg = reactionTweets
	}
	if count["news"] > 0 {
		reactionNews = sums["news"] / float32(count["news"])
		reactionAvg = reactionNews
	}
	if count["twitter"] > 0 && count["news"] > 0 {
		reactionAvg = (sums["twitter"] + sums["news"]) / float32(count["twitter"]+count["news"])
	}
	return reactionAvg, reactionTweets, reactionNews
}

func analyzeText(client *language.Client, ctx context.Context, t text, c chan analyzedText, wg *sync.WaitGroup) {
	defer wg.Done()
	s, err := analyzeSentiment(client, ctx, t.text)
	if err != nil {
		log.Error(fmt.Errorf("Failed on call to analyzeSentiment, %v", err))
		return
	}
	c <- analyzedText{s, t.textProvider, t.timestamp}

}
func analyzeSentiment(client *language.Client, ctx context.Context, text string) (float32, error) {
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
		return 0, fmt.Errorf("Failed on analyzing sentiment for %v, %v", text, err)
	}
	return sentiment.DocumentSentiment.Score, nil
}
