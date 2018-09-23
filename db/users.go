package db

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"
)

type User struct {
	id       int
	username string
	email    string
	hash     string
	token    string
}

func (env Env) getUsersWithName(username string) ([]User, error) {
	users := []User{}
	rows, err := env.db.Query("SELECT * FROM users where username=?", username)
	if err != nil {
		return nil, fmt.Errorf("failed on selecting users with name %v, %v", username, err)
	}
	defer rows.Close()
	for i := 0; rows.Next(); i++ {
		user := User{}
		if err := rows.Scan(&user.id, &user.username, &user.email, &user.hash, &user.token); err != nil {
			return nil, fmt.Errorf("Rows scan failed in getUsersWithName %v, %v", username, err)
		}
		users = append(users, user)
	}
	log.Debug(users)
	return users, nil
}

func (env Env) getUserWithName(username string) (User, error) {
	users, err := env.getUsersWithName(username)
	if err != nil {
		return User{}, fmt.Errorf("failed on call to getUsersWithName %v, %v", username, err)
	}
	l := len(users)
	if l != 1 {
		return User{}, fmt.Errorf("amount of users with this name %v is equal to %v", username, l)
	}
	return users[0], nil
}

func (env Env) UserIsPresent(username string) (bool, error) {
	users, err := env.getUsersWithName(username)
	if err != nil {
		return false, fmt.Errorf("failed on call to getUsersWithName %v, %v", username, err)
	}
	if len(users) != 0 {
		return true, nil
	}
	return false, nil
}

func (env Env) CreateUser(username, email, password string) (string, error) {
	token := newToken()
	_, err := env.db.Exec("INSERT INTO users (username, email, hash, token) VALUES (?, ?, ?, ?)", username, email, getSHA1Hash(password+env.salt), token)
	if err != nil {
		return "", fmt.Errorf("failed on inserting user %v, %v", username, err)
	}
	return token, nil

}
func (env Env) PasswordIsCorrect(username, password string) (bool, error) {
	user, err := env.getUserWithName(username)
	if err != nil {
		return false, err
	}
	if getSHA1Hash(password+env.salt) == user.hash {
		return true, nil
	}
	return false, nil

}
func (env Env) UpdateToken(username string) (string, error) {
	token := newToken()
	_, err := env.db.Exec("UPDATE users SET token=? WHERE username=?", token, username)
	return token, err

}
func (env Env) AuthenticateUser(username string, token string) (bool, error) {
	user, err := env.getUserWithName(username)
	if err != nil {
		return false, err
	}
	if user.token == token {
		return true, nil
	}
	return false, nil

}
func init() {
	rand.Seed(time.Now().UnixNano())
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func getSHA1Hash(text string) string {
	hasher := sha1.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func newToken() string {
	token := randSeq(32)
	return token
}
