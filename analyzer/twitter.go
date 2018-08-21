package analyzer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type twitterApi struct {
	Statuses []status `json: "statuses"`
}

type status struct {
	CreatedAt time.Time `json: "created_at"`
	Id        int       `json: "id"`
	Text      string    `json: "text"`
}

func getTweets(keyword, lang, date, twitterApiKey string) ([]text, error) {
	tt := []text{}
	client := clientWithTimeout(false)
	langParam := ""
	if lang != "any" {
		if lang == "gb" || lang == "us" {
			lang = "en"
		}
		langParam = fmt.Sprintf("&lang=%v", lang)
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.twitter.com/1.1/search/tweets.json?q=%v%v", keyword, langParam), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", twitterApiKey)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var tA twitterApi
	err = decoder.Decode(&tA)
	if err != nil {
		return nil, err
	}
	for _, s := range tA.Statuses {
		t := text{
			id:        s.Id,
			text:      s.Text,
			timestamp: s.CreatedAt,
		}
		log.Debug(t)
		tt = append(tt, t)
	}
	return tt, nil
}
