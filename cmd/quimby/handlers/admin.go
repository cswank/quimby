package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/cswank/quimby"
)

func GetClients(w http.ResponseWriter, req *http.Request) {
	json.NewEncoder(w).Encode(quimby.Clients)
}
