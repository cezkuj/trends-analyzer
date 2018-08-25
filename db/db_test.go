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

func TestTagIsNotPresent(t *testing.T) {
	env := setupEnv()
	present, err := env.tagIsPresent("trend")
	if err != nil {
		t.Fatal(err)
	}
	if present {
		t.Fatal(err)
	}
	cleanUp()
}

func TestTagIsPresent(t *testing.T) {
	env := setupEnv()
	tag := NewTag("trend", "", "")
	err := env.CreateTag(tag)
	if err != nil {
		t.Fatal(err)
	}
	present, err := env.tagIsPresent("trend")
	if err != nil {
		t.Fatal(err)
	}
	if !present {
		t.Fatal("Tag is not present")
	}
	cleanUp()
}

func TestGetTags(t *testing.T) {
	env := setupEnv()
	tag1 := NewTag("trends1", "", "")
	err := env.CreateTag(tag1)
	if err != nil {
		t.Fatal(err)
	}
	tags, err := env.GetTagsWithName(tag1.Name)
	if tags[0].Name != tag1.Name {
		t.Fatal("Tags name not equal")
	}
	cleanUp()
}

func TestGetTagID(t *testing.T) {
	env := setupEnv()
	tag1 := NewTag("trends1", "", "")
	tag2 := NewTag("trends2", "", "")
	err := env.CreateTag(tag1)
	if err != nil {
		t.Fatal(err)
	}
	err = env.CreateTag(tag2)
	if err != nil {
		t.Fatal(err)
	}
	id, err := env.GetTagID("trends2")
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
	tag1 := NewTag("trends1", "", "")
	err := env.CreateTag(tag1)
	if err != nil {
		t.Fatal(err)
	}
	tag_id, err := env.GetTagID(tag1.Name)
	if err != nil {
		t.Fatal(err)
	}
	a1 := NewAnalyzis(
		tag_id,
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
		tag_id,
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
	analyzes, err := env.GetAnalyzes(tag1.Name, time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC), time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC))
	if a1 != analyzes[0] || a2 != analyzes[1] {
		t.Fatal("Wrong analyzes in 2000-2020 query")
	}
	analyzes, err = env.GetAnalyzes(tag1.Name, time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC), time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC))
	if len(analyzes) != 0 {
		t.Fatal("Wrong analyzes in 2020-2020 query")
	}
	analyzes, err = env.GetAnalyzes(tag1.Name, time.Date(2011, 1, 1, 12, 0, 0, 0, time.UTC), time.Date(2016, 1, 1, 12, 0, 0, 0, time.UTC))
	if a2 != analyzes[0] {
		t.Fatal("Wrong analyzes in 2011-2016 query")
	}
	analyzes, err = env.GetAnalyzes(tag1.Name, time.Date(2009, 1, 1, 12, 0, 0, 0, time.UTC), time.Date(2012, 1, 1, 12, 0, 0, 0, time.UTC))
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
	truncateTable("tags")
	truncateTable("analyzes")
	log.SetLevel(log.DebugLevel)
	return Env{db: db}

}

func cleanUp() {
	truncateTable("tags")
	truncateTable("analyzes")

}
