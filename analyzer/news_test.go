package analyzer

import (
	"net/http"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
)

type mockClient struct {
	file string
}

func (c mockClient) Do(req *http.Request) (*http.Response, error) {
	file, err := os.Open(c.file)
	if err != nil {
		log.Fatal(err)
	}
	return &http.Response{Body: file}, nil
}

func TestGetNews(t *testing.T) {
	c := apiClient{NewsAPIUrl, "", mockClient{"examples/news.json"}}
	nn, err := c.getNews("Trump", "any", "any")
	log.Info(nn)
	if err != nil {
		t.Fatal(err)
	}
}
