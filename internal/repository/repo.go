package repository

import (
	"fmt"

	"github.com/asdine/storm"
	"github.com/cswank/quimby/internal/schema"
)

func New(pth string) (*Gadget, *User, error) {
	db, err := storm.Open(pth)
	if err != nil {
		return nil, nil, fmt.Errorf("could not open database: %s", err)
	}

	if err := db.Init(&schema.Gadget{}); err != nil {
		return nil, nil, fmt.Errorf("could not Init db: %v", err)
	}

	if err := db.Init(&schema.User{}); err != nil {
		return nil, nil, fmt.Errorf("could not Init db: %v", err)
	}

	return &Gadget{db: db}, &User{db: db}, nil
}
