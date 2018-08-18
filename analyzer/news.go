package analyzer

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type newsApi struct {
	Status       string    `json:"status"`
	TotalResults int       `json:"totalResults"`
	Articles     []article `json:"articles"`
}

type article struct {
	Title       string    `json: "title"`
	Description string    `json: "description"`
	PublishedAt time.Time `json: "publishedAt"`
}

func AnalyzeNews(keyword, country, date, newsApiKey string) {
	nn, err := getNews(keyword, country, date, newsApiKey)
	if err != nil {
		log.Error(err)
		return
	}
	for _, n := range nn {
		go analyzeText(n, "news")
	}
}

func getNews(keyword, country, date, newsApiKey string) ([]text, error) {
	tt := []text{}
	client := clientWithTimeout(false)
	countryParam := ""
	if country != "any" {
		countryParam = fmt.Sprintf("&country=%v", country)
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://newsapi.org/v2/top-headlines?q=%v%v", keyword, countryParam), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-Api-Key", newsApiKey)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var nA newsApi
	err = decoder.Decode(&nA)
	if err != nil {
		return nil, err
	}
	for _, a := range nA.Articles {
		txt := fmt.Sprintf("%v %v", a.Title, a.Description)
		id := hash(txt)
		t := text{
			id:        id,
			text:      txt,
			timestamp: a.PublishedAt,
		}
		log.Debug(t)
		tt = append(tt, t)
	}
	return tt, nil
}

func hash(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int(h.Sum32())
}
