package server

import (
	"log"
	"testing"
	"time"
)

const (
	DB_CONNECTION = "ta:trends_analyzer@tcp(127.0.0.1:3306)/trends"
)

func truncateTable(table_name string) {
	db, err := initDb(DB_CONNECTION)
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

}

func TestTagIsPresent(t *testing.T) {
	env := setupEnv()
	tag := NewTag("trend", "", "")
	err := env.createTag(tag)
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

}

func TestGetTags(t *testing.T) {
	env := setupEnv()
	tag1 := NewTag("trends1", "", "")
	err := env.createTag(tag1)
	if err != nil {
		t.Fatal(err)
	}
	tags, err := env.getTags(tag1.Name)
	if tags[0].Name != tag1.Name {
		t.Fatal("Tags name not equal")
	}

}

func TestGetTagID(t *testing.T) {
	env := setupEnv()
	tag1 := NewTag("trends1", "", "")
	tag2 := NewTag("trends2", "", "")
	err := env.createTag(tag1)
	if err != nil {
		t.Fatal(err)
	}
	err = env.createTag(tag2)
	if err != nil {
		t.Fatal(err)
	}
	id, err := env.getTagID("trends2")
	if err != nil {
		t.Fatal(err)
	}
	if id != 2 {
		t.Fatal("ID has unexpected value")
	}

}

func TestGetAnalyzes(t *testing.T) {
	env := setupEnv()
	tag1 := NewTag("trends1", "", "")
	err := env.createTag(tag1)
	if err != nil {
		t.Fatal(err)
	}
	tag_id, err := env.getTagID(tag1.Name)
	if err != nil {
		t.Fatal(err)
	}
	a1 := NewAnalyzis(
		tag_id,
		time.Date(2009, 1, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2012, 1, 1, 12, 0, 0, 0, time.UTC),
                0,
                0,
		float64(0.0),
		float64(0.0),
		float64(0.0),
	)
	err = env.createAnalyzis(a1)
	if err != nil {
		t.Fatal(err)
	}
	a2 := NewAnalyzis(
		tag_id,
		time.Date(2013, 1, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2015, 1, 1, 12, 0, 0, 0, time.UTC),
                0,
                0,
		float64(0.0),
		float64(0.0),
		float64(0.0),
	)
	err = env.createAnalyzis(a2)
	if err != nil {
		t.Fatal(err)
	}
        analyzes, err := env.getAnalyzes(tag1.Name, time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC), time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC))
        log.Println(analyzes)
}
func setupEnv() Env {
	db, err := initDb(DB_CONNECTION)
	if err != nil {
		log.Fatal(err)
	}
	truncateTable("tags")
	truncateTable("analyzes")
	return Env{db: db}

}
