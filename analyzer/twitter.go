package analyzer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	TwitterAPIUrl = "https://api.twitter.com"
)

type twitterAPI struct {
	Statuses []status `json:"statuses"`
}

type twitterClient struct {
	twitterAPIUrl string
	twitterAPIKey string
	httpClient    http.Client
}

type status struct {
	//json decoding does not work for anything else than RFC 3339 format - decoding to string first
	CreatedAt string `json:"created_at"`
	ID        int    `json:"id"`
	Text      string `json:"text"`
}

func (c twitterClient) getTweets(keyword, lang, date string) ([]text, error) {
	tt := []text{}
	langParam := ""
	if lang != "any" {
		if lang == "gb" || lang == "us" {
			lang = "en"
		}
		langParam = fmt.Sprintf("&lang=%v", lang)
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("%v/1.1/search/tweets.json?q=%v%v&count=100", c.twitterAPIUrl, keyword, langParam), nil)
	if err != nil {
		return nil, fmt.Errorf("Failed on creating twitter request in getTweets, %v", err)
	}
	req.Header.Add("Authorization", c.twitterAPIKey)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed on executing %v in getTweets, %v", req, err)
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var tA twitterAPI
	err = decoder.Decode(&tA)
	if err != nil {
		return nil, fmt.Errorf("Failed on decoding %v in getTweets, %v", resp.Body, err)
	}
	for _, s := range tA.Statuses {
		parsedTimestamp, err := time.Parse(time.RubyDate, s.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("Failed on parsing %v in getTweets, %v", time.RubyDate, err)
		}
		t := text{
			id:           s.ID,
			text:         s.Text,
			timestamp:    parsedTimestamp,
			textProvider: "twitter",
		}
		log.Debug(t)
		tt = append(tt, t)
	}
	return tt, nil
}
