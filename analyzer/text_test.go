package analyzer

import (
	log "github.com/sirupsen/logrus"
	"testing"

	language "cloud.google.com/go/language/apiv1"
	"golang.org/x/net/context"
)

func TestAnalyzeSentiment(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	ctx := context.Background()
	client, err := language.NewClient(ctx)
	if err != nil {
		t.Fatal(err)
	}
	sent, err := analyzeSentiment(client, ctx, "I am happy.")
	if err != nil {
		t.Fatal(err)
	}
	log.Debug(sent)
}
