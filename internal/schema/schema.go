package schema

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cswank/gogadgets"
)

// Gadget represents a gadget
type Gadget struct {
	ID     int    `storm:"id,increment"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	status map[string]gogadgets.Message
}

func (g *Gadget) Status() map[string]gogadgets.Message {
	return g.status
}

// FetchStatus queries the gadget to get its current
func (g *Gadget) FetchStatus() error {
	resp, err := http.Get(fmt.Sprintf("%s/gadgets", g.URL))
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(&g.status)
}
