package server

import (
	"errors"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Env struct {
	db *sqlx.DB
}

type Analyzis struct {
	ID             int
	TagID          int
	TimestampFirst time.Time
	TimestampLast  time.Time
	AmountOfTweets int
	AmountOfNews   int
	ReactionAvg    float64
	ReactionTweets float64
	ReactionNews   float64
}

type Tag struct {
	ID             int
	Name           string
	Provider       string
	AdditionalInfo string
}

func initDb(db_connection string) (*sqlx.DB, error) {
	db, err := sqlx.Open("mysql",
		db_connection+"parseTime=true")

	if err != nil {
		return nil, err
	}
	createAnalyses := `
          CREATE TABLE IF NOT EXISTS analyses (
          id SERIAL NOT NULL PRIMARY KEY,
          tag_id INT NOT NULL,
          timestamp_first DATETIME NOT NULL,
          timestamp_last DATETIME NOT NULL,
          amount_of_tweets INT NOT NULL,
          amount_of_news INT NOT NULL,
          reaction_avg FLOAT NOT NULL,
          reaction_tweets FLOAT NOT NULL,
          reaction_news FLOAT NOT NULL);
        `
	_, err = db.Exec(createAnalyses)
	if err != nil {
		return nil, err
	}
	createTags := `
          CREATE TABLE IF NOT EXISTS tags (
          id SERIAL NOT NULL PRIMARY KEY,
          name TEXT NOT NULL,
          provider TEXT NOT NULL,
          additional_info TEXT NOT NULL);
        `
	_, err = db.Exec(createTags)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (env Env) createTag(tag Tag) error {
	tPresent, err := env.tagIsPresent(tag.Name)
	if err != nil {
		return err
	}
	if tPresent {
		return errors.New("Tag already present")
	}
	_, err = env.db.Exec("INSERT INTO tags (name, provider, additional_info) VALUES (?, ?, ?)", tag.Name, tag.Provider, tag.AdditionalInfo)
	return err
}

func (env Env) getTagID(name string) (int, error) {
	tags := []Tag{}
	err := env.db.Select(&tags, "SELECT * FROM tags where name=?", name)
	if err != nil {
		return 0, nil
	}
	if len(tags) != 0 {
		return tags[0].ID, nil
	}
	return 0, errors.New("There is no tag with that name")

}

func (env Env) tagIsPresent(name string) (bool, error) {
	tags := []Tag{}
	err := env.db.Select(&tags, "SELECT * FROM tags where name=?", name)
	if len(tags) != 0 {
		return true, err
	}
	return false, err
}

func (env Env) createAnalyzis(a Analyzis, tagName string) error {
	tagID, err := env.getTagID(tagName)
	if err != nil {
		return err
	}
	_, err = env.db.Exec("INSERT INTO analyzes (tag_id, timestamp_first, timestamp_last, amount_of_tweets, amount_of_news, reaction_avg, reaction_tweets, reaction_news) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", tagID, a.TimestampFirst, a.TimestampLast, a.AmountOfTweets, a.AmountOfNews, a.ReactionAvg, a.ReactionTweets, a.ReactionNews)
	return err
}

func (env Env) getAnalyzes(tagName string, timestampFirst, timestampLast time.Time) ([]Analyzis, error) {
	tagID, err := env.getTagID(tagName)
	if err != nil {
		return nil, err
	}
	analyzes := []Analyzis{}
	err = env.db.Select(&analyzes, "SELECT * FROM analyzes where tag_id=? AND timestamp_first >=? AND timestamp_last <=?", tagID, timestampFirst, timestampLast)
	if err != nil {
		return nil, err
	}
	return analyzes, nil

}
