package controllers

import (
	"encoding/json"

	"github.com/cswank/quimby/models"
)

func GetClients(args *Args) error {
	d, err := json.Marshal(models.Clients)
	args.W.Write(d)
	return err
}
