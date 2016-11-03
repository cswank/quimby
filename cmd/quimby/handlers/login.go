package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cswank/quimby"
)

var (
	exp = time.Duration(24 * time.Hour)
)

func Logout(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:   "quimby",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
	args := GetArgs(r)
	if args.Args.Get("web") == "true" {
		w.Header().Set("Location", "/login.html")
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	user := quimby.NewUser("", quimby.UserDB(DB), quimby.UserTFA(TFA))
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(user)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := doLogin(user, w, r); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
	}
}

func doLogin(user *quimby.User, w http.ResponseWriter, req *http.Request) error {
	goodPassword, err := user.CheckPassword()
	if !goodPassword || err != nil {
		return fmt.Errorf("bad request")
	}

	params := req.URL.Query()
	methods, ok := params["auth"]
	user.TFAData = []byte{}
	if ok && methods[0] == "jwt" {
		setToken(w, user)
	} else {
		setCookie(w, user)
	}
	return nil
}

func setToken(w http.ResponseWriter, user *quimby.User) {
	token, err := quimby.GenerateToken(user.Username, exp)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
	} else {

		w.Header().Set("Authorization", token)
	}
}

func setCookie(w http.ResponseWriter, user *quimby.User) {
	http.SetCookie(w, quimby.GenerateCookie(user.Username))
}
