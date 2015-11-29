package models

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
	"github.com/cswank/gogadgets"
)

type Gadget struct {
	Id   string   `json:"id"`
	Name string   `json:"name"`
	Host string   `json:"host"`
	DB   *bolt.DB `json:"-"`
}

var (
	NotFound = errors.New("not found")
	_gadgets = []byte("gadgets")
	_notes   = []byte("notes")
	_stats   = []byte("stats")
	epoch    = []byte(time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339Nano))
	century  = (100 * 24 * 365 * time.Hour)
)

func GetGadgets(db *bolt.DB) ([]Gadget, error) {
	gadgets := []Gadget{}

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(_gadgets)
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			g := Gadget{DB: db}
			if err := json.Unmarshal(v, &g); err != nil {
				return err
			}
			gadgets = append(gadgets, g)
		}
		return nil
	})
	return gadgets, err
}

func (g *Gadget) Fetch() error {
	return g.DB.View(func(tx *bolt.Tx) error {
		v := tx.Bucket(_gadgets).Get([]byte(g.Id))
		if len(v) == 0 {
			return NotFound
		}
		return json.Unmarshal(v, g)
	})
}

type DataPoint struct {
	Time  time.Time `json:"time"`
	Value float64   `json:"value"`
}

func (g *Gadget) SaveDataPoint(name string, dp DataPoint) error {
	return g.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.Bucket([]byte(g.Id)).Bucket(_stats).CreateBucketIfNotExists([]byte(name))
		if err != nil {
			return nil
		}
		buf := new(bytes.Buffer)
		if err := binary.Write(buf, binary.LittleEndian, dp.Value); err != nil {
			return err
		}
		return b.Put([]byte(dp.Time.Format(time.RFC3339Nano)), buf.Bytes())
	})
}

func (g *Gadget) GetDataPoints(name string, start, end time.Time) ([]DataPoint, error) {
	var points []DataPoint
	err := g.DB.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(g.Id)).Bucket(_stats).Bucket([]byte(name)).Cursor()
		min := []byte(start.Format(time.RFC3339Nano))
		max := []byte(end.Format(time.RFC3339Nano))
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			var val float64
			buf := bytes.NewReader(v)
			if err := binary.Read(buf, binary.LittleEndian, &val); err != nil {
				return err
			}
			ts, _ := time.Parse(time.RFC3339Nano, string(k))
			points = append(points, DataPoint{Time: ts, Value: val})
		}
		return nil
	})
	return points, err
}

func (g *Gadget) Save() error {
	return g.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(_gadgets)
		if g.Id == "" {
			if err := g.createGadget(tx); err != nil {
				return err
			}
		}
		d, _ := json.Marshal(g)
		return b.Put([]byte(g.Id), d)
	})
}

func (g *Gadget) createGadget(tx *bolt.Tx) error {
	g.Id = gogadgets.GetUUID()
	b, err := tx.CreateBucket([]byte(g.Id))
	if err != nil {
		return err
	}
	if _, err := b.CreateBucket(_notes); err != nil {
		return err
	}
	_, err = b.CreateBucket(_stats)
	return err
}

type Note struct {
	Time   string `json:"time,omitempty"`
	Text   string `json:"text"`
	Author string `json:"author"`
}

func (g *Gadget) AddNote(note Note) error {
	return g.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(g.Id)).Bucket(_notes)
		d, _ := json.Marshal(note)
		return b.Put([]byte(time.Now().Format(time.RFC3339Nano)), d)
	})
}

func (g *Gadget) GetNotes(start, end *time.Time) ([]Note, error) {
	var notes []Note
	g.DB.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(g.Id)).Bucket(_notes).Cursor()
		min, max := g.getMinMax(start, end)
		for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			var n Note
			n.Time = string(k)
			json.Unmarshal(v, &n)
			notes = append(notes, n)
		}
		return nil
	})
	return notes, nil
}

func (g *Gadget) getMinMax(start, end *time.Time) ([]byte, []byte) {
	var min, max []byte
	if start == nil {
		min = epoch
	} else {
		min = []byte(start.Format(time.RFC3339Nano))
	}
	if end == nil {
		max = []byte(time.Now().Add(century).Format(time.RFC3339Nano))
	} else {
		max = []byte(end.Format(time.RFC3339Nano))
	}
	return min, max
}

func (g *Gadget) Delete() error {
	return g.DB.Update(func(tx *bolt.Tx) error {
		if err := tx.DeleteBucket([]byte(g.Id)); err != nil {
			return err
		}
		b := tx.Bucket(_gadgets)
		return b.Delete([]byte(g.Id))
	})
}

func (g *Gadget) Update(cmd string) error {
	m := gogadgets.Message{
		UUID:   gogadgets.GetUUID(),
		Sender: "quimby",
		Type:   gogadgets.COMMAND,
		Body:   cmd,
	}
	return g.UpdateMessage(m)
}

func (g *Gadget) ReadDevice(w io.Writer, location, device string) error {
	r, err := http.Get(fmt.Sprintf("%s/gadgets/locations/%s/devices/%s/status", g.Host, location, device))
	if err != nil {
		return err
	}
	defer r.Body.Close()
	_, err = io.Copy(w, r.Body)
	return err
}

func (g *Gadget) GetDevice(location string, name string) (gogadgets.Value, error) {
	m, err := g.GetValues()
	v, ok := m[location][name]
	if !ok {
		return v, fmt.Errorf("%s %s not found", location, name)
	}
	return v, err
}

func (g *Gadget) UpdateDevice(location string, name string, v gogadgets.Value) error {
	cmd := g.getCommand(location, name, v)
	return g.SendCommand(cmd)
}

func (g *Gadget) SendCommand(cmd string) error {
	m := gogadgets.Message{
		UUID:   gogadgets.GetUUID(),
		Sender: "quimby",
		Type:   gogadgets.COMMAND,
		Body:   cmd,
	}
	return g.UpdateMessage(m)
}

func (g *Gadget) getCommand(location string, name string, v gogadgets.Value) string {
	return fmt.Sprintf("%s %s %s", g.getVerb(v), location, name)
}

func (g *Gadget) getVerb(v gogadgets.Value) string {
	if v.Value == true {
		return "turn on"
	}
	return "turn off"
}

func (g *Gadget) UpdateMessage(m gogadgets.Message) error {
	buf := bytes.Buffer{}
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(m); err != nil {
		return err
	}
	r, err := http.Post(fmt.Sprintf("%s/gadgets", g.Host), "application/json", &buf)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response from %s: %d", g.Host, r.StatusCode)
	}
	return nil
}

func (g *Gadget) Status() (map[string]gogadgets.Message, error) {
	var m map[string]gogadgets.Message
	r, err := http.Get(fmt.Sprintf("%s/gadgets", g.Host))

	if err != nil {
		return m, err
	}
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	return m, dec.Decode(&m)
}

func (g *Gadget) ReadStatus(w io.Writer) error {
	r, err := http.Get(fmt.Sprintf("%s/gadgets", g.Host))
	if err != nil {
		return err
	}
	defer r.Body.Close()
	_, err = io.Copy(w, r.Body)
	return err
}

func (g *Gadget) GetValues() (map[string]map[string]gogadgets.Value, error) {
	r, err := http.Get(fmt.Sprintf("%s/gadgets/values", g.Host))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	var m map[string]map[string]gogadgets.Value
	dec := json.NewDecoder(r.Body)
	return m, dec.Decode(&m)
}

func (g *Gadget) ReadValues(w io.Writer) error {
	r, err := http.Get(fmt.Sprintf("%s/gadgets/values", g.Host))
	if err != nil {
		return err
	}
	defer r.Body.Close()
	_, err = io.Copy(w, r.Body)
	return err
}

func (g *Gadget) Register(addr, token string) (string, error) {
	m := map[string]string{"address": addr, "token": token}

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.Encode(&m)
	r, err := http.Post(fmt.Sprintf("%s/clients", g.Host), "application/json", buf)
	if err != nil {
		return "", err
	}
	r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected response from %s: %d", g.Host, r.StatusCode)
	}
	return g.Host, nil
}
