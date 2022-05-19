package server

import (
	"log"

	"github.com/cswank/quimby/internal/auth"
	"github.com/cswank/quimby/internal/config"
	"github.com/cswank/quimby/internal/homekit"
	"github.com/cswank/quimby/internal/repository"
	"github.com/cswank/quimby/internal/router"
)

func Start(cfg config.Config, g *repository.Gadget, u *repository.User) {
	a := auth.New(u)
	hc, err := homekit.New()
	if err != nil {
		log.Fatal(err)
	}

	if err := router.Serve(cfg, g, u, a, hc); err != nil {
		log.Fatal(err)
	}
}
