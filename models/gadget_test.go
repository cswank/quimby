package models_test

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/boltdb/bolt"
	. "github.com/cswank/quimby/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Gadgets", func() {
	var (
		g    *Gadget
		dir  string
		pth  string
		db   *bolt.DB
		host string
	)

	BeforeEach(func() {
		var err error
		host = "111.222.333.444"
		dir, err = ioutil.TempDir("", "")
		pth = path.Join(dir, "db")
		Expect(err).To(BeNil())

		db, err = GetDB(pth)
		Expect(err).To(BeNil())

		g = &Gadget{
			Name: "lights",
			Host: host,
			DB:   db,
		}
		err = g.Save()
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		db.Close()
		os.RemoveAll(dir)
	})

	It("can save", func() {
		g2 := &Gadget{
			Name: "lights",
			DB:   db,
		}
		err := g2.Fetch()
		Expect(err).To(BeNil())
		Expect(g2.Host).To(Equal(host))
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
		Expect(g2.Host).To(Equal(host))
	})
})
