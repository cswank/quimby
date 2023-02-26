package schema

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Gadget represents a gadget
type Gadget struct {
	ID     int    `storm:"id,increment"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	Status map[string]map[string]Message
}

func (g Gadget) String() string {
	return fmt.Sprintf("%d: %s %s", g.ID, g.Name, g.URL)
}

func (g *Gadget) Register(addr, token string) (string, error) {
	m := map[string]string{"address": addr, "token": token}

	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(&m)
	if err != nil {
		return "", err
	}

	r, err := http.Post(fmt.Sprintf("%s/clients", g.URL), "application/json", buf)
	if err != nil {
		return "", err
	}

	r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected response from %s: %d", g.URL, r.StatusCode)
	}

	return g.URL, nil
}

func (g Gadget) Send(m Message) error {
	buf := bytes.Buffer{}
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(m); err != nil {
		return err
	}
	resp, err := http.Post(fmt.Sprintf("%s/gadgets", g.URL), "application/json", &buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response from %s: %d", g.URL, resp.StatusCode)
	}
	return nil
}

func (g *Gadget) Command(cmd string) error {
	return g.Send(Message{
		UUID:   UUID(),
		Sender: "quimby",
		Type:   "command",
		Body:   cmd,
	})
}

// Fetch queries the gadget to get its current status
func (g *Gadget) Fetch() error {
	resp, err := http.Get(fmt.Sprintf("%s/gadgets", g.URL))
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	var m map[string]Message
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return err
	}

	status := map[string]map[string]Message{}

	for _, v := range m {
		if v.Name == "" || v.Location == "" {
			continue
		}

		l, ok := status[v.Location]
		if !ok {
			l = map[string]Message{}
		}

		l[v.Name] = v
		status[v.Location] = l
	}

	g.Status = status
	return nil
}

// UUID generates a random UUID according to RFC 4122
func UUID() string {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return ""
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}

type Message struct {
	UUID        string    `json:"uuid"`
	From        string    `json:"from,omitempty"`
	Name        string    `json:"name,omitempty"`
	Location    string    `json:"location,omitempty"`
	Type        string    `json:"type,omitempty"`
	Sender      string    `json:"sender,omitempty"`
	Target      string    `json:"target,omitempty"`
	Body        string    `json:"body,omitempty"`
	Host        string    `json:"host,omitempty"`
	Method      Method    `json:"method,omitempty"`
	Timestamp   time.Time `json:"timestamp,omitempty"`
	Value       Value     `json:"value,omitempty"`
	TargetValue *Value    `json:"target_value,omitempty"`
	Info        struct {
		Direction string   `json:"direction"`
		Type      string   `json:"type"`
		On        []string `json:"on"`
		Off       []string `json:"off"`
	} `json:"info"`
}

type Method struct {
	Step  int      `json:"step,omitempty"`
	Steps []string `json:"steps,omitempty"`
	Time  int      `json:"time,omitempty"`
}

type Value struct {
	Value    interface{}     `json:"value,omitempty"`
	Units    string          `json:"units,omitempty"`
	Output   map[string]bool `json:"io,omitempty"`
	ID       string          `json:"id,omitempty"`
	Cmd      string          `json:"command,omitempty"`
	location string
	name     string
}

func (v *Value) GetName() string {
	return v.name
}

func (v *Value) ToFloat() (f float64, ok bool) {
	switch V := v.Value.(type) {
	case bool:
		if V {
			f = 1.0
		} else {
			f = 0.0
		}
		ok = true
	case float64:
		f = V
		ok = true
	}
	return f, ok
}
