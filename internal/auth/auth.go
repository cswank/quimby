package auth

import (
	"crypto"
	"errors"
	"net/http"
	"time"

	"github.com/cswank/quimby/internal/config"
	repo "github.com/cswank/quimby/internal/repository"
	"github.com/cswank/quimby/internal/schema"
	"github.com/gorilla/securecookie"
	"github.com/sec51/twofactor"
	"golang.org/x/crypto/bcrypt"
)

const (
	week = 24 * 7 * time.Hour
)

type Auth struct {
	repo *repo.User
	sc   *securecookie.SecureCookie
}

func New(r *repo.User) *Auth {
	cfg := config.Get()
	return &Auth{
		repo: r,
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
		Expires:  time.Now().Add(week),
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

func Credentials(name, pws string) ([]byte, []byte, []byte, error) {
	pw, err := bcrypt.GenerateFromPassword([]byte(pws), 10)
	if err != nil {
		return nil, nil, nil, err
	}

	t, qr, err := tfa(name)
	return pw, t, qr, err
}

func tfa(username string) ([]byte, []byte, error) {
	otp, err := twofactor.NewTOTP(username, "quimby", crypto.SHA1, 6)
	if err != nil {
		return nil, nil, err
	}

	data, err := otp.ToBytes()
	if err != nil {
		return nil, nil, err
	}

	qr, err := otp.QR()
	return data, qr, err
}
