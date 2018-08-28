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
	KeywordID          int       `json:"keyword_id"`
	Country        string    `json:"country"`
	Timestamp      time.Time `json:"timestamp"`
	AmountOfTweets int       `json:"amount_of_tweets"`
	AmountOfNews   int       `json:"amount_of_news"`
	ReactionAvg    float32   `json:"reaction_avg"`
	ReactionTweets float32   `json:"reaction_tweets"`
	ReactionNews   float32   `json:"reaction_news"`
}

func NewAnalyzis(keywordID int, country string, timestamp time.Time, amountOfTweets, amountOfNews int, reactionAvg, reactionTweets, reactionNews float32) Analyzis {
	return Analyzis{keywordID, country, timestamp, amountOfTweets, amountOfNews, reactionAvg, reactionTweets, reactionNews}
}

type Keyword struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Provider       string `json:"provider"`
	AdditionalInfo string `json:"additional_info"`
}

func NewKeyword(name, provider, additionalInfo string) Keyword {
	return Keyword{0, name, provider, additionalInfo}
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
          keyword_id INT NOT NULL,
          country TEXT NOT NULL,
          timestamp DATETIME NOT NULL,
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
	createKeywords := `
          CREATE TABLE IF NOT EXISTS keywords (
          id SERIAL NOT NULL PRIMARY KEY,
          name TEXT NOT NULL,
          provider TEXT NOT NULL,
          additional_info TEXT NOT NULL);
        `
	_, err = db.Exec(createKeywords)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (env Env) CreateKeyword(keyword Keyword) error {
	tPresent, err := env.keywordIsPresent(keyword.Name)
	if err != nil {
		return err
	}
	if tPresent {
		return errors.New("Keyword already present")
	}
	_, err = env.db.Exec("INSERT INTO keywords (name, provider, additional_info) VALUES (?, ?, ?)", keyword.Name, keyword.Provider, keyword.AdditionalInfo)
	if err != nil {
		return err
	}
	log.Debug("Keyword " + keyword.Name + " inserted")
	return nil
}

func (env Env) CreateKeywordIfNotPresent(keyword Keyword) error {
	tPresent, err := env.keywordIsPresent(keyword.Name)
	if err != nil {
		return err
	}
	if tPresent {
		return nil
	}
	err = env.CreateKeyword(keyword)
	return err

}

func (env Env) GetKeywordID(name string) (int, error) {
	keywords, err := env.GetKeywordsWithName(name)
	if err != nil {
		return -1, err
	}
	if len(keywords) != 1 {
		return -1, errors.New("Keyword does not exist")
	}
	return keywords[0].ID, nil

}
func (env Env) GetKeywordsWithName(name string) ([]Keyword, error) {
	return env.getKeywords("SELECT * FROM keywords where name=?", name)
}

func (env Env) GetKeywords() ([]Keyword, error) {
	return env.getKeywords("SELECT * FROM keywords")
}

func (env Env) getKeywords(query string, args ...interface{}) ([]Keyword, error) {
	keywords := []Keyword{}
	rows, err := env.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for i := 0; rows.Next(); i++ {
		keyword := Keyword{}
		if err := rows.Scan(&keyword.ID, &keyword.Name, &keyword.Provider, &keyword.AdditionalInfo); err != nil {
			return nil, err
		}
		keywords = append(keywords, keyword)
	}
	log.Debug(keywords)
	return keywords, nil
}

func (env Env) keywordIsPresent(name string) (bool, error) {
	keywords, err := env.GetKeywordsWithName(name)
	if err != nil {
		return false, err
	}
	if len(keywords) != 0 {
		return true, nil
	}
	return false, nil
}

func (env Env) CreateAnalyzis(a Analyzis) error {
	_, err := env.db.Exec("INSERT INTO analyzes (keyword_id, country, timestamp, amount_of_tweets, amount_of_news, reaction_avg, reaction_tweets, reaction_news) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", a.KeywordID, a.Country, a.Timestamp, a.AmountOfTweets, a.AmountOfNews, a.ReactionAvg, a.ReactionTweets, a.ReactionNews)
	if err != nil {
		return err
	}
	log.Debug(a)
	return nil
}

func (env Env) getAnalyzes(query string, args ...interface{}) ([]Analyzis, error) {
	analyzes := []Analyzis{}
	rows, err := env.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for i := 0; rows.Next(); i++ {
		a := Analyzis{}
		if err := rows.Scan(&a.KeywordID, &a.Country, &a.Timestamp, &a.AmountOfTweets, &a.AmountOfNews, &a.ReactionAvg, &a.ReactionTweets, &a.ReactionNews); err != nil {
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

func (env Env) GetAnalyzes(keywordName string, after, before time.Time, country string) ([]Analyzis, error) {
	keywordID, err := env.GetKeywordID(keywordName)
	if err != nil {
		return nil, err
	}
	if country == "any" {
		return env.getAnalyzes("SELECT keyword_id, country, timestamp, amount_of_tweets, amount_of_news, reaction_avg, reaction_tweets, reaction_news FROM analyzes WHERE keyword_id=? AND timestamp >=? AND timestamp <=?", keywordID, after, before)
	}
	return env.getAnalyzes("SELECT keyword_id, country, timestamp, amount_of_tweets, amount_of_news, reaction_avg, reaction_tweets, reaction_news FROM analyzes WHERE keyword_id=? AND timestamp >=? AND timestamp <=? AND country=?", keywordID, after, before, country)

}
