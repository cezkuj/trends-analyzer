package analyzer

import (
	log "github.com/sirupsen/logrus"
	"testing"
)

func TestAnalyzeSentiment(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	sent, err := analyzeSentiment("I am happy.")
	if err != nil {
		t.Fatal(err)
	}
	log.Debug(sent)
}
