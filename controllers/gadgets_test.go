
import (
	"io/ioutil"
	"os"
	"path"

	"github.com/boltdb/bolt"
)

var _ = Describe("Gadgets", func() {
	var (
		lights     *Gadget
		sprinklers *Gadget
		dir        string
		pth        string
		db         *bolt.DB
	)

	BeforeEach(func() {
		var err error
		dir, err = ioutil.TempDir("", "")
		pth = path.Join(dir, "db")
		Expect(err).To(BeNil())

		db, err = GetDB(pth)
		Expect(err).To(BeNil())

		lights = &Gadget{
			Name: "lights",
			Host: ts.URL,
			DB:   db,
		}

		sprinlers = &Gadget{
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
})
