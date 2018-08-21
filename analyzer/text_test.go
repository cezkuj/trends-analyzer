package analyzer

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestAnalyzeText(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	analyzeText(text{}, "")
}
