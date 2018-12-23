package userhttp

import (
	"net/http"

	"github.com/cswank/quimby/internal/middleware"
	"github.com/go-chi/chi"
)

func New(r chi.Router) {
	u := &UserHTTP{}
	r.Get("/admin/users", middleware.Handle(u.GetAll))
}

type UserHTTP struct {
}

func (u UserHTTP) GetAll(w http.ResponseWriter, req *http.Request) error {
	_, err := w.Write([]byte("all users"))
	return err
}
