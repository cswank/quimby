package models_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"

	"github.com/boltdb/bolt"
	. "github.com/cswank/quimby/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Gadgets", func() {
	var (
		g   *Gadget
		dir string
		pth string
		db  *bolt.DB
		ts  *httptest.Server
	)

	BeforeEach(func() {

		ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		Expect(len(gadgets.Gadgets)).To(Equal(1))
		g2 := gadgets.Gadgets[0]
		Expect(g2.Host).To(Equal(ts.URL))
	})

	It("gets the status of the gadget", func() {
		status, err := g.Status()
		Expect(err).To(BeNil())
		Expect(len(status)).To(Equal(5))
		v := status["back yard"]["sprinklers"]
		Expect(v.Value).To(BeFalse())
	})
})
