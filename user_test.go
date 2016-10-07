package quimby_test

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/boltdb/bolt"
	. "github.com/cswank/quimby"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Users", func() {
	var (
		u   *User
		dir string
		pth string
		db  *bolt.DB
	)

	BeforeEach(func() {
		var err error
		dir, err = ioutil.TempDir("", "")
		pth = path.Join(dir, "db")
		Expect(err).To(BeNil())

		db, err = GetDB(pth)
		Expect(err).To(BeNil())

		u = NewUser(
			"inspector",
			UserPassword("abc123bs"),
			UserPermission("write"),
			UserDB(db),
		)
		err = u.Save()
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		db.Close()
		os.RemoveAll(dir)
	})

	It("can save", func() {
		u2 := NewUser("inspector", UserDB(db))
		Expect(u2.Fetch()).To(BeNil())
		Expect(u2.Password).To(Equal(""))
		Expect(len(u2.HashedPassword)).ToNot(Equal(0))
		Expect(u2.Permission).To(Equal("write"))
	})

	It("can delete", func() {
		err := u.Delete()
		Expect(err).To(BeNil())
		u2 := NewUser("inspector", UserDB(db))
		err = u2.Fetch()
		Expect(err).ToNot(BeNil())
		Expect(u2.Permission).To(Equal(""))
	})

	It("is authorized", func() {
		u2 := NewUser("inspector", UserDB(db))

		Expect(u2.IsAuthorized("write")).To(BeTrue())
	})

	It("is not authorized", func() {
		u2 := NewUser("inspector", UserDB(db))
		Expect(u2.IsAuthorized("admin")).To(BeFalse())
	})

	It("is not authorized after a delete", func() {
		u.Delete()
		u2 := NewUser("inspector", UserDB(db))
		Expect(u2.IsAuthorized("write")).To(BeFalse())
	})

	It("checks its password", func() {
		u2 := NewUser("inspector", UserDB(db), UserPassword("abc123bs"))
		good, err := u2.CheckPassword()
		Expect(err).To(BeNil())
		Expect(good).To(BeTrue())
	})

	It("checks a deleted password", func() {
		u.Delete()
		u2 := NewUser("inspector", UserDB(db), UserPassword("abc123bs"))
		good, err := u2.CheckPassword()
		Expect(err).ToNot(BeNil())
		Expect(good).To(BeFalse())
	})
})
