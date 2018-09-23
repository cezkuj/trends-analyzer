package db

import (
	"testing"
)

func TestGetUsersWithName(t *testing.T) {
	env := setupEnv()
	users, err := env.getUsersWithName("abc")
	if err != nil {
		t.Error(err)
	}
	if len(users) != 0 {
		t.Error("There should be no users with name abc")
	}
	_, err = env.CreateUser("abc", "abc", "abc")
	if err != nil {
		t.Error(err)
	}
	users, err = env.getUsersWithName("abc")
	if err != nil {
		t.Error(err)
	}
	if len(users) != 1 {
		t.Error("User abc not found")
	}
	users, err = env.getUsersWithName("abc2")
	if err != nil {
		t.Error(err)
	}
	if len(users) != 0 {
		t.Error("There should be no users with name abc2")
	}
}

func TestGetUserWithName(t *testing.T) {
	env := setupEnv()
	_, err := env.getUserWithName("abc")
	if err == nil {
		t.Error("There should be error in case user is not present")
	}
	_, err = env.CreateUser("abc3", "abc3", "abc3")
	if err != nil {
		t.Error(err)
	}
	user, err := env.getUserWithName("abc3")
	if user.username != "abc3" {
		t.Error("User has invalid name")
	}
}

func TestUserIsPresent(t *testing.T) {
	env := setupEnv()
	present, err := env.UserIsPresent("abc4")
	if err != nil {
		t.Error(err)
	}
	if present {
		t.Error("User abc4 should not be present")
	}
	_, err = env.CreateUser("abc4", "abc4", "abc4")
	if err != nil {
		t.Error(err)
	}
	present, err = env.UserIsPresent("abc4")
	if err != nil {
		t.Error(err)
	}
	if !present {
		t.Error("User abc4 should be present")
	}
}

func TestCreateUser(t *testing.T) {
	env := setupEnv()
	_, err := env.CreateUser("abc5", "abc5", "abc5")
	if err != nil {
		t.Error(err)
	}
	present, err := env.UserIsPresent("abc5")
	if err != nil || !present {
		t.Error("abc5 user is not present")
	}
	user, err := env.getUserWithName("abc5")
	if err != nil {
		t.Error(err)
	}
	if user.email != "abc5" {
		t.Errorf("%v user has invalid email", user)
	}
}

func TestPasswordIsCorrect(t *testing.T) {
	env := setupEnv()
	_, err := env.CreateUser("abc6", "abc6", "abc6")
	if err != nil {
		t.Error(err)
	}
	correct, err := env.PasswordIsCorrect("abc6", "abc7")
	if err != nil {
		t.Error(err)
	}
	if correct {
		t.Error("Password should not be correct")
	}
	correct, err = env.PasswordIsCorrect("abc6", "abc6")
	if err != nil {
		t.Error(err)
	}
	if !correct {
		t.Error("Password should be correct")
	}
}

func TestUpdateToken(t *testing.T) {
	env := setupEnv()
	token, err := env.CreateUser("abc7", "abc7", "abc7")
	if err != nil {
		t.Error(err)
	}
	token2, err := env.UpdateToken("abc7")
	if err != nil {
		t.Error(err)
	}
	if token == token2 {
		t.Error("Token should be updateed")
	}
}

func TestAuthenticateUser(t *testing.T) {
	env := setupEnv()
	token, err := env.CreateUser("abc8", "abc8", "abc8")
	if err != nil {
		t.Error(err)
	}
	authenticated, err := env.AuthenticateUser("abc8", "abc8")
	if err != nil {
		t.Error(err)
	}
	if authenticated {
		t.Error("User should not be authenticated")
	}
	authenticated, err = env.AuthenticateUser("abc8", token)
	if err != nil {
		t.Error(err)
	}
	if !authenticated {
		t.Error("User should be authenticated")
	}
}
