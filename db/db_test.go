package db

import (
	log "github.com/sirupsen/logrus"
	"testing"
	"time"
)

const (
	DB_CONNECTION = "ta:trends_analyzer@tcp(127.0.0.1:3306)/trends"
)

func truncateTable(table_name string) {
	db, err := InitDb(DB_CONNECTION)
	if err != nil {
		log.Fatal(err)
	}
	db.Exec("TRUNCATE TABLE " + table_name)
}

func TestKeywordIsNotPresent(t *testing.T) {
	env := setupEnv()
	present, err := env.keywordIsPresent("trend")
	if err != nil {
		t.Fatal(err)
	}
	if present {
		t.Fatal(err)
	}
	cleanUp()
}

func TestKeywordIsPresent(t *testing.T) {
	env := setupEnv()
	keyword := NewKeyword("trend", "", "")
	err := env.CreateKeyword(keyword)
	if err != nil {
		t.Fatal(err)
	}
	present, err := env.keywordIsPresent("trend")
	if err != nil {
		t.Fatal(err)
	}
	if !present {
		t.Fatal("Keyword is not present")
	}
	cleanUp()
}

func TestGetKeywords(t *testing.T) {
	env := setupEnv()
	keyword1 := NewKeyword("trends1", "", "")
	err := env.CreateKeyword(keyword1)
	if err != nil {
		t.Fatal(err)
	}
	keywords, err := env.GetKeywordsWithName(keyword1.Name)
	if keywords[0].Name != keyword1.Name {
		t.Fatal("Keywords name not equal")
	}
	cleanUp()
}

func TestGetKeywordID(t *testing.T) {
	env := setupEnv()
	keyword1 := NewKeyword("trends1", "", "")
	keyword2 := NewKeyword("trends2", "", "")
	err := env.CreateKeyword(keyword1)
	if err != nil {
		t.Fatal(err)
	}
	err = env.CreateKeyword(keyword2)
	if err != nil {
		t.Fatal(err)
	}
	id, err := env.GetKeywordID("trends2")
	if err != nil {
		t.Fatal(err)
	}
	if id != 2 {
		t.Fatal("ID has unexpected value")
	}
	cleanUp()

}

func TestGetAnalyzes(t *testing.T) {
	env := setupEnv()
	keyword1 := NewKeyword("trends1", "", "")
	err := env.CreateKeyword(keyword1)
	if err != nil {
		t.Fatal(err)
	}
	keyword_id, err := env.GetKeywordID(keyword1.Name)
	if err != nil {
		t.Fatal(err)
	}
	a1 := NewAnalyzis(
		keyword_id,
		"us",
		time.Date(2009, 1, 1, 12, 0, 0, 0, time.UTC),
		0,
		0,
		float32(0.0),
		float32(0.0),
		float32(0.0),
	)
	err = env.CreateAnalyzis(a1)
	if err != nil {
		t.Fatal(err)
	}
	a2 := NewAnalyzis(
		keyword_id,
		"us",
		time.Date(2013, 1, 1, 12, 0, 0, 0, time.UTC),
		0,
		0,
		float32(0.0),
		float32(0.0),
		float32(0.0),
	)
	err = env.CreateAnalyzis(a2)
	if err != nil {
		t.Fatal(err)
	}
	analyzes, err := env.GetAnalyzes(keyword1.Name, time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC), time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC), "us")
	if a1 != analyzes[0] || a2 != analyzes[1] {
		t.Fatal("Wrong analyzes in 2000-2020 query")
	}
	analyzes, err = env.GetAnalyzes(keyword1.Name, time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC), time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC), "us")
	if len(analyzes) != 0 {
		t.Fatal("Wrong analyzes in 2020-2020 query")
	}
	analyzes, err = env.GetAnalyzes(keyword1.Name, time.Date(2011, 1, 1, 12, 0, 0, 0, time.UTC), time.Date(2016, 1, 1, 12, 0, 0, 0, time.UTC), "us")
	if a2 != analyzes[0] {
		t.Fatal("Wrong analyzes in 2011-2016 query")
	}
	analyzes, err = env.GetAnalyzes(keyword1.Name, time.Date(2009, 1, 1, 12, 0, 0, 0, time.UTC), time.Date(2012, 1, 1, 12, 0, 0, 0, time.UTC), "us")
	if a1 != analyzes[0] {
		t.Fatal("Wrong analyzes in 2009-2012 query")
	}
	cleanUp()

}
func setupEnv() Env {
	db, err := InitDb(DB_CONNECTION)
	if err != nil {
		log.Fatal(err)
	}
	truncateTable("keywords")
	truncateTable("analyzes")
	log.SetLevel(log.DebugLevel)
	return Env{db: db}

}

func cleanUp() {
	truncateTable("keywords")
	truncateTable("analyzes")

}
