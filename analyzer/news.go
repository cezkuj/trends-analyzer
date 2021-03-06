package analyzer

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	NewsAPIUrl = "https://newsapi.org"
)

type newsAPI struct {
	Status       string    `json:"status"`
	TotalResults int       `json:"totalResults"`
	Articles     []article `json:"articles"`
}

type article struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	PublishedAt time.Time `json:"publishedAt"`
}

func (c apiClient) getNews(keyword, country, date string) ([]text, error) {
	tt := []text{}
	countryParam := ""
	if country != "any" {
		countryParam = fmt.Sprintf("&country=%v", country)
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("%v/v2/top-headlines?q=%v%v", c.apiUrl, keyword, countryParam), nil)
	if err != nil {
		return nil, fmt.Errorf("Failed on creating requests in getNews, %v", err)
	}
	req.Header.Add("X-Api-Key", c.apiKey)
	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed on executing %v in getNews", req)
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var nA newsAPI
	err = decoder.Decode(&nA)
	if err != nil {
		return nil, fmt.Errorf("Failed on decoding %v, in getNews, %v", resp.Body, err)
	}
	for _, a := range nA.Articles {
		txt := fmt.Sprintf("%v %v", a.Title, a.Description)
		id := hash(txt)
		t := text{
			id:           id,
			text:         txt,
			timestamp:    a.PublishedAt,
			textProvider: "news",
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
