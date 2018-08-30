package analyzer

import (
	"time"
        "fmt"

        log "github.com/sirupsen/logrus"

	"github.com/cezkuj/trends-analyzer/db"
)

func StartDispatching(env db.Env) {
	keywords, err := env.GetKeywords()
	if err != nil {
		log.Error(err)
		return
	}
	for {
		for _, k := range keywords {
                        log.Info(fmt.Sprintf("Started analyzing: %v", k))
			go Analyze(env, k.Name, "both", "us", "any")
			time.Sleep(5 * time.Second)
		}

	}
}
