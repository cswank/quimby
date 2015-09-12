package controllers_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"

	"github.com/boltdb/bolt"
	"github.com/cswank/gogadgets"
	"github.com/cswank/quimby/controllers"
	"github.com/cswank/quimby/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Gadgets", func() {
	var (
		lights     *models.Gadget
		sprinklers *models.Gadget
		dir        string
		pth        string
		db         *bolt.DB
		args       *controllers.Args
		ts         *httptest.Server
		msgs       []gogadgets.Message
		clients    []map[string]string
		w          *httptest.ResponseRecorder
	)

	BeforeEach(func() {

		msgs = []gogadgets.Message{}
		clients = []map[string]string{}
		dir, _ = ioutil.TempDir("", "")
		pth = path.Join(dir, "db")

		var err error
		db, err = models.GetDB(pth)
		Expect(err).To(BeNil())

		w = httptest.NewRecorder()

		args = &controllers.Args{
			W:  w,
			R:  &http.Request{},
			DB: db,
		}

		ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" {
				fmt.Fprintln(
					w,
					`{
  "kitchen": {
    "lights": {
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

		lights = &models.Gadget{
			Name: "lights",
			Host: ts.URL,
			DB:   db,
		}
		err = lights.Save()
		Expect(err).To(BeNil())

		sprinklers = &models.Gadget{
			Name: "sprinklers",
			Host: ts.URL,
			DB:   db,
		}
		err = sprinklers.Save()
		Expect(err).To(BeNil())

	})

	AfterEach(func() {
		db.Close()
		os.RemoveAll(dir)
		ts.Close()
	})

	Context("gadgets", func() {
		It("returns all gadgets in the system", func() {
			err := controllers.GetGadgets(args)
			Expect(err).To(BeNil())
			var gadgets []models.Gadget
			dec := json.NewDecoder(w.Body)
			err = dec.Decode(&gadgets)
			Expect(err).To(BeNil())
			Expect(len(gadgets)).To(Equal(2))
			g := gadgets[0]
			Expect(g.Name).To(Equal("lights"))
			g2 := gadgets[1]
			Expect(g2.Name).To(Equal("sprinklers"))
		})
	})
})
