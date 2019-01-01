package middleware

import (
	"errors"
	"net/http"

	"github.com/cswank/quimby/internal/config"
	"github.com/cswank/quimby/internal/schema"
	"github.com/cswank/quimby/internal/user"
	"github.com/cswank/quimby/internal/user/repository"
	"github.com/gorilla/securecookie"
)

type Auth struct {
	repo user.Repository
	sc   *securecookie.SecureCookie
}

func NewAuth() *Auth {
	cfg := config.Get()
	return &Auth{
		repo: repository.New(),
		sc:   securecookie.New([]byte(cfg.HashKey), []byte(cfg.BlockKey)),
	}
}

func (a *Auth) Auth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		u, err := a.getUserFromCookie(req)
		if err != nil || u == nil {
			http.Redirect(w, req, "/login", http.StatusSeeOther)
		} else {
			h.ServeHTTP(w, req)
		}
	})
}

func (a *Auth) GenerateCookie(username string) (*http.Cookie, error) {
	value := map[string]string{
		"username": username,
	}

	encoded, err := a.sc.Encode("quimby", value)
	return &http.Cookie{
		Name:     "quimby",
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
	}, err
}

func (a *Auth) getUserFromCookie(r *http.Request) (*schema.User, error) {
	cookie, err := r.Cookie("quimby")
	if err != nil {
		return nil, err
	}

	var m map[string]string
	err = a.sc.Decode("quimby", cookie.Value, &m)
	if err != nil {
		return nil, err
	}

	un, ok := m["username"]
	if !ok || un == "" {
		return nil, errors.New("no way, eh")
	}

	return &schema.User{
		Name: un,
	}, nil
}
