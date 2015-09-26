package controllers

import (
	"net/http"

	"github.com/cswank/quimby/models"
)

func getUserFromCookie(r *http.Request) (*models.User, error) {
	user := &models.User{
		DB: DB,
	}
	cookie, err := r.Cookie("quimby")
	if err != nil {
		return nil, err
	}
	var m map[string]string
	err = sc.Decode("quimby", cookie.Value, &m)
	if err != nil {
		return nil, err
	}
	user.Username = m["user"]
	err = user.Fetch()
	user.HashedPassword = []byte{}
	return user, err
}
