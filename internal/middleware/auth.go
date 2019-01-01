package middleware

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/cswank/quimby/internal/user"
	"github.com/cswank/quimby/internal/user/repository"
)

type Auth struct {
	repo user.Repository
}

func NewAuth() *Auth {
	return &Auth{
		repo: repository.New(),
	}
}

func (a *Auth) Auth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		u, err := a.repo.Get(1)
		fmt.Println(u, err)
		if err != nil {
			http.Redirect(w, req, "/login", http.StatusSeeOther)
		} else {
			h.ServeHTTP(w, req)
		}
	})
}

func GenerateCookie(username string) *http.Cookie {
	value := map[string]string{
		"user": username,
	}

	encoded, _ := SC.Encode("quimby", value)
	return &http.Cookie{
		Name:     "quimby",
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
	}
}

func GetUserFromCookie(r *http.Request) (*User, error) {
	user := NewUser("")
	cookie, err := r.Cookie("quimby")

	if err != nil {
		return nil, err
	}
	var m map[string]string
	err = SC.Decode("quimby", cookie.Value, &m)
	if err != nil {
		return nil, err
	}
	if m["user"] == "" {
		return nil, errors.New("no way, eh")
	}
	user.Username = m["user"]
	err = user.Fetch()
	user.HashedPassword = []byte{}
	user.TFAData = []byte{}
	return user, err
}
