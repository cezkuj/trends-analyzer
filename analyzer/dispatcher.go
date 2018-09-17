package analyzer

import (
	"fmt"
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/cezkuj/trends-analyzer/db"
)

func StartDispatching(env db.Env, interval int) {
	rand.Seed(time.Now().Unix())
	for {
		keywords, err := env.GetKeywords()
		if err != nil {
			log.Error(fmt.Errorf("GetKeywords in StartDispatching failed on %v", err))
			return
		}
		k := keywords[rand.Intn(len(keywords))]
		a, err := env.GetAnalyzes(k.Name, time.Time{}, time.Now(), "any")
		if err != nil {
			log.Error(fmt.Errorf("GetAnalyzes in StartDispatching for %v failed on %v", k, err))
			return
		}
		log.Info(fmt.Sprintf("Started analyzing: %v", k))
		go Analyze(env, k.Name, "both", a[0].Country, "any")
		time.Sleep(time.Duration(interval) * time.Minute)

	}
}
