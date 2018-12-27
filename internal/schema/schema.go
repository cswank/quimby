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
	status map[string]map[string]gogadgets.Message
}

func (g *Gadget) Status() map[string]map[string]gogadgets.Message {
	return g.status
}

// FetchStatus queries the gadget to get its current
func (g *Gadget) FetchStatus() error {
	resp, err := http.Get(fmt.Sprintf("%s/gadgets", g.URL))
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	var m map[string]gogadgets.Message
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return err
	}

	status := map[string]map[string]gogadgets.Message{}

	for _, v := range m {
		if v.Name == "" || v.Location == "" {
			continue
		}

		l, ok := status[v.Location]
		if !ok {
			l = map[string]gogadgets.Message{}
		}

		l[v.Name] = v
		status[v.Location] = l
	}

	g.status = status
	return nil
}
