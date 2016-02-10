package handlers

import (
	"encoding/json"

	"github.com/cswank/quimby"
)

func GetClients(args *Args) error {
	d, err := json.Marshal(quimby.Clients)
	args.W.Write(d)
	return err
}
