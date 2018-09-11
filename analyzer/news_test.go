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
	if err != nil {
		t.Fatal(err)
	}
	expected := text{id: 2182534426,
		text:         "Trump administration to sanction International Criminal Court, ban judges from US Human rights advocates and others  bashed the sanctions against the ICC and decision to close Palestinian office in Washington.",
		textProvider: "news"}
	if nn[0].id != expected.id || nn[0].text != expected.text || nn[0].textProvider != expected.textProvider {
		t.Fatalf("%v is not equal to %v", nn[0], expected)
	}
}
