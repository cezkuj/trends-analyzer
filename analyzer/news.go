package analyzer

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type news struct {
	id        int
	text      string
	timestamp time.Time
}

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

func (n news) Text() string {
	return n.text
}

func AnalyzeNews(keyword, country, date, newsApiKey string) {
	nn, err := getNews(keyword, country, date, newsApiKey)
	if err != nil {
		log.Error(err)
		return
	}
	for _, n := range nn {
		log.Debug(fmt.Sprintf("Analyzing news %v - %v", n.id, n.timestamp))
		go analyzeText(n, "news")
	}
}

func getNews(keyword, country, date, newsApiKey string) ([]news, error) {
	nn := []news{}
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
		text := fmt.Sprintf("%v %v", a.Title, a.Description)
		id := hash(text)
		n := news{
			id:        id,
			text:      text,
			timestamp: a.PublishedAt,
		}
		log.Debug(n)
		nn = append(nn, n)
	}
	return nn, nil
}

func hash(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int(h.Sum32())
}
