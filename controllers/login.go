package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/cswank/quimby/models"
)

func Logout(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:   "quimby",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
}

func Login(w http.ResponseWriter, r *http.Request) {
	user := &models.User{
		DB: DB,
	}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(user)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	goodPassword, err := user.CheckPassword()
	if !goodPassword {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	value := map[string]string{
		"user": user.Username,
	}

	encoded, _ := sc.Encode("quimby", value)
	cookie := &http.Cookie{
		Name:     "quimby",
		Value:    encoded,
		Path:     "/",
		HttpOnly: false,
	}
	w.Header().Set("Location", "/api/users/current")
	http.SetCookie(w, cookie)
}
