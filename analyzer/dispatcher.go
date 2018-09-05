package analyzer

import (
	"fmt"
	"time"

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
			a, err := env.GetAnalyzes(k.Name, time.Time{}, time.Now(), "any")
			if err != nil {
				log.Error(err)
				return
			}
			log.Info(fmt.Sprintf("Started analyzing: %v", k))
			go Analyze(env, k.Name, "both", a[0].Country, "any")
			time.Sleep(3 * time.Minute)
		}

	}
}
