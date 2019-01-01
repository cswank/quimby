package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/cswank/quimby/internal/schema"
	"github.com/cswank/quimby/internal/user"
	"github.com/cswank/quimby/internal/user/repository"
	"github.com/gorilla/securecookie"
)

var (
	hashKey  = []byte(os.Getenv("QUIMBY_HASH_KEY"))
	blockKey = []byte(os.Getenv("QUIMBY_BLOCK_KEY"))
	sc       = securecookie.New(hashKey, blockKey)
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
		u, err := getUserFromCookie(req)
		if err != nil || u == nil {
			http.Redirect(w, req, "/login", http.StatusSeeOther)
		} else {
			h.ServeHTTP(w, req)
		}
	})
}

func GenerateCookie(username string) (*http.Cookie, error) {
	if len(hashKey) != 16 || len(blockKey) != 16 {
		return nil, fmt.Errorf("you must set QUIMBY_HASH_KEY and QUIMBY_BLOCK_KEY, 16 characters each")
	}

	value := map[string]string{
		"username": username,
	}

	encoded, err := sc.Encode("quimby", value)
	return &http.Cookie{
		Name:     "quimby",
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
	}, err
}

func getUserFromCookie(r *http.Request) (*schema.User, error) {
	cookie, err := r.Cookie("quimby")

	if err != nil {
		return nil, err
	}

	var m map[string]string
	err = sc.Decode("quimby", cookie.Value, &m)
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
