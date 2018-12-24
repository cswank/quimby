package middleware

import (
	"log"
	"net/http"

	"github.com/cswank/quimby/internal/templates"
)

type Handler func(http.ResponseWriter, *http.Request) error
type RenderFunc func(http.ResponseWriter, *http.Request) (Renderer, error)

func Handle(h Handler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		err := h(w, req)
		if err != nil {
			log.Println(err)
		}
	}
}

func Render(r RenderFunc) func(w http.ResponseWriter, req *http.Request) error {
	return func(w http.ResponseWriter, req *http.Request) error {
		pg, err := r(w, req)
		if err != nil || pg == nil {
			return err
		}

		t, scripts, stylesheets := templates.Get(pg.Template())
		pg.AddScripts(scripts)
		pg.AddStylesheets(stylesheets)
		return t.ExecuteTemplate(w, "base", pg)
	}
}

// Renderer supplies the data needed to render html
type Renderer interface {
	AddScripts([]string)
	AddStylesheets([]string)
	Template() string
}
