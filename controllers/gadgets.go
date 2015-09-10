package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/cswank/quimby/models"
)

func GetGadgets(w http.ResponseWriter, r *http.Request, u *models.User, vars map[string]string) error {
	g, err := models.GetGadgets(db)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	return enc.Encode(g)
}

// func GetStatus(w http.ResponseWriter, r *http.Request, u *models.User, vars map[string]string) error {
// 	host, ok := vars["name"]
// 	if !ok {
// 		return errors.New("you must supply a host arg")
// 	}

// 	r, err := http.Get(fmt.Sprintf("%s/status/values", host))

// }
