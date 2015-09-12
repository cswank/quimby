package controllers

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/cswank/quimby/models"
)

func GetGadgets(args *Args) error {
	g, err := models.GetGadgets(args.DB)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(args.W)
	return enc.Encode(g)
}

func GetGadget(args *Args) error {
	err := args.Gadget.Fetch()
	if err != nil {
		return err
	}
	enc := json.NewEncoder(args.W)
	return enc.Encode(args.Gadget)
}

func DeleteGadget(args *Args) error {
	return args.Gadget.Delete()
}

func GetStatus(args *Args) error {
	if err := args.Gadget.Fetch(); err != nil {
		return err
	}

	return args.Gadget.Status(args.W)
}

func SendCommand(args *Args) error {
	if err := args.Gadget.Fetch(); err != nil {
		return err
	}

	var m map[string]string
	dec := json.NewDecoder(args.R.Body)
	if err := dec.Decode(&m); err != nil {
		return err
	}
	return args.Gadget.Update(m["command"])
}

func AddGadget(args *Args) error {
	var g models.Gadget
	dec := json.NewDecoder(args.R.Body)
	err := dec.Decode(&g)
	if err != nil {
		return err
	}
	g.DB = args.DB
	err = g.Save()
	if err != nil {
		return err
	}

	u, err := url.Parse(fmt.Sprintf("/api/gadgets/%s", g.Name))
	if err != nil {
		return err
	}
	args.W.Header().Set("Location", u.String())
	return nil
}

// func GetStatus(w http.ResponseWriter, r *http.Request, u *models.User, vars map[string]string) error {
// 	host, ok := vars["name"]
// 	if !ok {
// 		return errors.New("you must supply a host arg")
// 	}

// 	r, err := http.Get(fmt.Sprintf("%s/status/values", host))

// }
