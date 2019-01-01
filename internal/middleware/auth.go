package middleware

import (
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
