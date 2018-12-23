package gadgethttp

import (
	"net/http"

	"github.com/cswank/quimby/internal/gadget"
	"github.com/cswank/quimby/internal/middleware"
	"github.com/go-chi/chi"
)

func New(r chi.Router) {
	g := &GadgetHTTP{}
	r.Get("/gadgets", middleware.Handle(g.GetAll))
}

// GadgetHTTP renders html
type GadgetHTTP struct {
	repo gadget.Repository
}

// GetAll shows all the gadgets
func (g GadgetHTTP) GetAll(w http.ResponseWriter, req *http.Request) error {
	gadgets, err := g.repo.GetAll()
	if err != nil {
		return err
	}

	for _, ga := range gadgets {
		w.Write([]byte(ga.Name))
	}

	return nil
}
