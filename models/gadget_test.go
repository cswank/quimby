package models_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"

	"github.com/boltdb/bolt"
	"github.com/cswank/gogadgets"
	. "github.com/cswank/quimby/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Gadgets", func() {
	var (
		g       *Gadget
		dir     string
		pth     string
		db      *bolt.DB
		ts      *httptest.Server
		msgs    []gogadgets.Message
		clients []map[string]string
	)

	BeforeEach(func() {
		msgs = []gogadgets.Message{}
		clients = []map[string]string{}

		ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" {
				fmt.Fprintln(
					w,
					`{
  "back garden": {
    "sprinklers": {
      "value": false,
      "io": false
    }
  },
  "back yard": {
    "sprinklers": {
      "value": false,
      "io": false
    }
  },
  "front garden": {
    "sprinklers": {
      "value": true,
      "io": true
    }
  },
  "front yard": {
    "sprinklers": {
      "value": false,
      "io": false
    }
  },
  "sidewalk": {
    "sprinklers": {
      "value": false,
      "io": false
    }
  }
}`,
				)
			} else if r.Method == "POST" {
				if r.URL.Path == "/gadgets" {
					var m gogadgets.Message
					dec := json.NewDecoder(r.Body)
					err := dec.Decode(&m)
					Expect(err).To(BeNil())
					msgs = append(msgs, m)
				} else if r.URL.Path == "/clients" {
					var m map[string]string
					dec := json.NewDecoder(r.Body)
					err := dec.Decode(&m)
					Expect(err).To(BeNil())
					clients = append(clients, m)
				}
			}
		}))

		var err error
		dir, err = ioutil.TempDir("", "")
		pth = path.Join(dir, "db")
		Expect(err).To(BeNil())

		db, err = GetDB(pth)
		Expect(err).To(BeNil())

		g = &Gadget{
			Name: "lights",
			Host: ts.URL,
			DB:   db,
		}
		err = g.Save()
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		db.Close()
		os.RemoveAll(dir)
		ts.Close()
	})

	It("can save", func() {
		g2 := &Gadget{
			Name: "lights",
			DB:   db,
		}
		err := g2.Fetch()
		Expect(err).To(BeNil())
		Expect(g2.Host).To(Equal(ts.URL))
	})

	It("can delete", func() {
		err := g.Delete()
		Expect(err).To(BeNil())

		g2 := &Gadget{
			Name: "lights",
			DB:   db,
		}
		err = g2.Fetch()
		Expect(err).ToNot(BeNil())
		Expect(g2.Host).To(Equal(""))
	})

	It("gets all gadgets", func() {
		gadgets, err := GetGadgets(db)
		Expect(err).To(BeNil())
		Expect(len(gadgets)).To(Equal(1))
		g2 := gadgets[0]
		Expect(g2.Host).To(Equal(ts.URL))
	})

	It("reads the status of the gadget", func() {
		buf := &bytes.Buffer{}
		err := g.ReadStatus(buf)
		Expect(err).To(BeNil())

		var status map[string]map[string]gogadgets.Value
		dec := json.NewDecoder(buf)
		err = dec.Decode(&status)
		Expect(err).To(BeNil())

		Expect(len(status)).To(Equal(5))
		v := status["back yard"]["sprinklers"]
		Expect(v.Value).To(BeFalse())
	})

	It("gets the status of the gadget", func() {
		status, err := g.Status()
		Expect(err).To(BeNil())

		Expect(len(status)).To(Equal(5))
		v := status["back yard"]["sprinklers"]
		Expect(v.Value).To(BeFalse())
	})

	It("updates the status of the gadget", func() {
		err := g.Update("turn on back yard sprinklers")
		Expect(err).To(BeNil())
		Expect(len(msgs)).To(Equal(1))
		m := msgs[0]
		Expect(m.Body).To(Equal("turn on back yard sprinklers"))
	})

	It("registers with a gogadgets instance", func() {
		h, err := g.Register(ts.URL, "fakecookie")
		Expect(err).To(BeNil())
		Expect(h).To(Equal(ts.URL))
		Expect(len(clients)).To(Equal(1))
		c := clients[0]
		Expect(c["address"]).To(Equal(ts.URL))
	})
})
