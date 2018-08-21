package db

import (
	"database/sql"
	"errors"
	"time"

	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
)

type Env struct {
	db            *sql.DB
	TwitterApiKey string
	NewsApiKey    string
}

func NewEnv(db *sql.DB, twitterApiKey, newsApiKey string) Env {
	return Env{db, twitterApiKey, newsApiKey}
}

type Analyzis struct {
	TagID          int
	TimestampFirst time.Time
	TimestampLast  time.Time
	AmountOfTweets int
	AmountOfNews   int
	ReactionAvg    float32
	ReactionTweets float32
	ReactionNews   float32
}
func NewAnalyzis(tagID int, timestampFirst, timestampLast time.Time, amountOfTweets, amountOfNews int, reactionAvg, reactionTweets, reactionNews float32) Analyzis {
	return Analyzis{tagID, timestampFirst, timestampLast, amountOfTweets, amountOfNews, reactionAvg, reactionTweets, reactionNews}
}

type Tag struct {
	ID             int
	Name           string
	Provider       string
	AdditionalInfo string
}

func NewTag(name, provider, additionalInfo string) Tag {
	return Tag{0, name, provider, additionalInfo}
}
func InitDb(db_connection string) (*sql.DB, error) {
	db, err := sql.Open("mysql",
		db_connection+"?parseTime=true")

	if err != nil {
		return nil, err
	}
	createAnalyzes := `
          CREATE TABLE IF NOT EXISTS analyzes (
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
	_, err = db.Exec(createAnalyzes)
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

func (env Env) CreateTag(tag Tag) error {
	tPresent, err := env.tagIsPresent(tag.Name)
	if err != nil {
		return err
	}
	if tPresent {
		return errors.New("Tag already present")
	}
	_, err = env.db.Exec("INSERT INTO tags (name, provider, additional_info) VALUES (?, ?, ?)", tag.Name, tag.Provider, tag.AdditionalInfo)
	if err != nil {
		return err
	}
	log.Debug("Tag " + tag.Name + " inserted")
	return nil
}

func (env Env) CreateTagIfNotPresent(tag Tag) error {
	tPresent, err := env.tagIsPresent(tag.Name)
	if err != nil {
		return err
	}
	if tPresent {
		return nil
	}
	err = env.CreateTag(tag)
	return err

}

func (env Env) GetTagID(name string) (int, error) {
	tags, err := env.getTags(name)
	if err != nil {
		return -1, err
	}
	if len(tags) != 1 {
		return -1, errors.New("Tag does not exist")
	}
	return tags[0].ID, nil

}
func (env Env) getTags(name string) ([]Tag, error) {
	tags := []Tag{}
	rows, err := env.db.Query("SELECT * FROM tags where name=?", name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for i := 0; rows.Next(); i++ {
		tag := Tag{}
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Provider, &tag.AdditionalInfo); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	log.Debug(tags)
	return tags, nil

}
func (env Env) tagIsPresent(name string) (bool, error) {
	tags, err := env.getTags(name)
	if err != nil {
		return false, err
	}
	if len(tags) != 0 {
		return true, nil
	}
	return false, nil
}

func (env Env) createAnalyzis(a Analyzis) error {
	_, err := env.db.Exec("INSERT INTO analyzes (tag_id, timestamp_first, timestamp_last, amount_of_tweets, amount_of_news, reaction_avg, reaction_tweets, reaction_news) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", a.TagID, a.TimestampFirst, a.TimestampLast, a.AmountOfTweets, a.AmountOfNews, a.ReactionAvg, a.ReactionTweets, a.ReactionNews)
	if err != nil {
		return err
	}
	log.Debug(a)
	return nil
}

func (env Env) getAnalyzes(tagName string, timestampFirst, timestampLast time.Time) ([]Analyzis, error) {
	tagID, err := env.GetTagID(tagName)
	if err != nil {
		return nil, err
	}
	analyzes := []Analyzis{}
	rows, err := env.db.Query("SELECT tag_id, timestamp_first, timestamp_last, amount_of_tweets, amount_of_news, reaction_avg, reaction_tweets, reaction_news FROM analyzes WHERE tag_id=? AND timestamp_first >=? AND timestamp_last <=?", tagID, timestampFirst, timestampLast)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for i := 0; rows.Next(); i++ {
		a := Analyzis{}
		if err := rows.Scan(&a.TagID, &a.TimestampFirst, &a.TimestampLast, &a.AmountOfTweets, &a.AmountOfNews, &a.ReactionAvg, &a.ReactionTweets, &a.ReactionNews); err != nil {
			return nil, err
		}
		analyzes = append(analyzes, a)
	}

	if err != nil {
		return nil, err
	}
	log.Debug(analyzes)
	return analyzes, nil
}
