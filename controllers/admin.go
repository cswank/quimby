package controllers

import "encoding/json"

func GetClients(args *Args) error {
	d, err := json.Marshal(Clients)
	args.W.Write(d)
	return err
}
