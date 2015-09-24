package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/cswank/gogadgets"
	"github.com/cswank/quimby/models"
	"github.com/gorilla/websocket"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

var cstDialer = websocket.Dialer{
	Subprotocols:    []string{"p1", "p2"},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type fakeLogger struct {
	f bool
}

func (f *fakeLogger) Println(v ...interface{})          {}
func (f *fakeLogger) Printf(s string, v ...interface{}) {}
func (f *fakeLogger) Fatal(v ...interface{})            { f.f = true }
func (f *fakeLogger) Fatalf(s string, v ...interface{}) { f.f = true }

var _ = Describe("Quimby", func() {
	var (
		port        string
		port2       string
		root        string
		iRoot       string
		dir         string
		pth         string
		u           *models.User
		u2          *models.User
		addr        string
		db          *bolt.DB
		cookies     []*http.Cookie
		readCookies []*http.Cookie
		lights      *models.Gadget
		sprinklers  *models.Gadget
		ts          *httptest.Server
		msgs        []gogadgets.Message
		clients     []map[string]string
		lg          *fakeLogger
	)

	BeforeEach(func() {
		lg = &fakeLogger{}
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
		port = fmt.Sprintf("%d", 1024+rand.Intn(65535-1024))
		port2 = fmt.Sprintf("%d", 1024+rand.Intn(65535-1024))

		root = fmt.Sprintf("%s", RandString(10))
		iRoot = fmt.Sprintf("%s", RandString(10))

		os.Setenv("QUIMBY_PORT", port)
		os.Setenv("QUIMBY_INTERNAL_PORT", port2)
		os.Setenv("QUIMBY_HOST", "http://localhost")
		addr = fmt.Sprintf("http://localhost:%s/api/%%s", port)

		dir, _ = ioutil.TempDir("", "")
		pth = path.Join(dir, "db")

		var err error
		db, err = models.GetDB(pth)
		Expect(err).To(BeNil())

		u = &models.User{
			Username:   "me",
			Password:   "hushhush",
			Permission: "write",
			DB:         db,
		}

		err = u.Save()
		Expect(err).To(BeNil())

		u2 = &models.User{
			Username:   "him",
			Password:   "shhhhhhhh",
			Permission: "read",
			DB:         db,
		}

		err = u2.Save()
		Expect(err).To(BeNil())

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

		go start(db, port, root, iRoot, lg)

		var r *http.Response
		Eventually(func() error {
			buf := bytes.Buffer{}
			enc := json.NewEncoder(&buf)
			usr := &models.User{
				Username: "me",
				Password: "hushhush",
			}
			enc.Encode(usr)
			url := fmt.Sprintf(addr, "login")
			var err error
			r, err = http.Post(url, "application/json", &buf)
			return err
		}).Should(BeNil())
		Expect(r.StatusCode).To(Equal(http.StatusOK))
		cookies = r.Cookies()

		buf := bytes.Buffer{}
		enc := json.NewEncoder(&buf)
		usr2 := &models.User{
			Username: "him",
			Password: "shhhhhhhh",
		}
		enc.Encode(usr2)
		url := fmt.Sprintf(addr, "login")
		r, err = http.Post(url, "application/json", &buf)
		Expect(err).To(BeNil())

		Expect(r.StatusCode).To(Equal(http.StatusOK))
		readCookies = r.Cookies()
	})

	AfterEach(func() {
		db.Close()
		os.RemoveAll(dir)
	})

	Context("logging in and out", func() {
		It("lets you log in", func() {
			Expect(len(cookies)).To(Equal(1))
		})
	})
	Context("not logged in", func() {
		It("does not let you get gadgets", func() {
			r, err := http.Get(fmt.Sprintf(addr, "gadgets"))
			Expect(err).To(BeNil())
			Expect(r.StatusCode).To(Equal(http.StatusUnauthorized))
		})
	})

	Context("logged in", func() {
		It("lets you get gadgets", func() {
			req, err := http.NewRequest("GET", fmt.Sprintf(addr, "gadgets"), nil)
			Expect(err).To(BeNil())
			req.AddCookie(cookies[0])
			r, err := http.DefaultClient.Do(req)
			Expect(err).To(BeNil())
			defer r.Body.Close()
			Expect(r.StatusCode).To(Equal(http.StatusOK))
			var gadgs []models.Gadget
			dec := json.NewDecoder(r.Body)
			err = dec.Decode(&gadgs)
			Expect(err).To(BeNil())

			Expect(len(gadgs)).To(Equal(2))
		})

		It("lets you get a gadget", func() {
			req, err := http.NewRequest("GET", fmt.Sprintf(addr, "gadgets/sprinklers"), nil)
			Expect(err).To(BeNil())
			req.AddCookie(cookies[0])
			r, err := http.DefaultClient.Do(req)
			Expect(err).To(BeNil())
			defer r.Body.Close()

			var g models.Gadget
			dec := json.NewDecoder(r.Body)
			err = dec.Decode(&g)
			Expect(err).To(BeNil())
			Expect(g.Name).To(Equal("sprinklers"))
			Expect(g.Host).To(Equal(ts.URL))
		})

		It("lets you delete a gadget", func() {
			req, err := http.NewRequest("DELETE", fmt.Sprintf(addr, "gadgets/sprinklers"), nil)
			Expect(err).To(BeNil())
			req.AddCookie(cookies[0])
			r, err := http.DefaultClient.Do(req)
			Expect(err).To(BeNil())
			defer r.Body.Close()

			req, err = http.NewRequest("GET", fmt.Sprintf(addr, "gadgets"), nil)
			Expect(err).To(BeNil())
			req.AddCookie(cookies[0])
			r, err = http.DefaultClient.Do(req)
			Expect(err).To(BeNil())
			defer r.Body.Close()
			Expect(r.StatusCode).To(Equal(http.StatusOK))
			var gadgs []models.Gadget
			dec := json.NewDecoder(r.Body)
			err = dec.Decode(&gadgs)
			Expect(err).To(BeNil())
			Expect(len(gadgs)).To(Equal(1))
		})

		It("doesn't let a read only user delete a gadget", func() {
			req, err := http.NewRequest("DELETE", fmt.Sprintf(addr, "gadgets/sprinklers"), nil)
			Expect(err).To(BeNil())
			req.AddCookie(readCookies[0])
			r, err := http.DefaultClient.Do(req)
			Expect(err).To(BeNil())
			Expect(r.StatusCode).To(Equal(http.StatusUnauthorized))
			defer r.Body.Close()

			req, err = http.NewRequest("GET", fmt.Sprintf(addr, "gadgets"), nil)
			Expect(err).To(BeNil())
			req.AddCookie(readCookies[0])
			r, err = http.DefaultClient.Do(req)
			Expect(err).To(BeNil())
			defer r.Body.Close()
			Expect(r.StatusCode).To(Equal(http.StatusOK))
			var gadgs []models.Gadget
			dec := json.NewDecoder(r.Body)
			err = dec.Decode(&gadgs)
			Expect(err).To(BeNil())
			Expect(len(gadgs)).To(Equal(2))
		})

		It("lets you add a gadget", func() {
			var buf bytes.Buffer
			enc := json.NewEncoder(&buf)
			g := models.Gadget{
				Name: "back yard",
				Host: "http://overthere.com",
			}
			enc.Encode(g)

			req, err := http.NewRequest("POST", fmt.Sprintf(addr, "gadgets"), &buf)
			Expect(err).To(BeNil())
			req.AddCookie(cookies[0])
			r, err := http.DefaultClient.Do(req)
			Expect(err).To(BeNil())
			defer r.Body.Close()
			Expect(r.StatusCode).To(Equal(http.StatusOK))
			Expect(len(r.Header["Location"])).To(Equal(1))
			u := r.Header["Location"][0]
			Expect(u).To(Equal("/api/gadgets/back%20yard"))

			req, err = http.NewRequest("GET", fmt.Sprintf(addr, "gadgets"), nil)
			Expect(err).To(BeNil())
			req.AddCookie(cookies[0])
			r, err = http.DefaultClient.Do(req)
			Expect(err).To(BeNil())
			defer r.Body.Close()
			Expect(r.StatusCode).To(Equal(http.StatusOK))
			var gadgs []models.Gadget
			dec := json.NewDecoder(r.Body)
			err = dec.Decode(&gadgs)
			Expect(err).To(BeNil())

			Expect(len(gadgs)).To(Equal(3))
		})

		It("lets you get the status of a gadget", func() {
			req, err := http.NewRequest("GET", fmt.Sprintf(addr, "gadgets/sprinklers/status"), nil)
			Expect(err).To(BeNil())
			req.AddCookie(readCookies[0])
			r, err := http.DefaultClient.Do(req)
			Expect(err).To(BeNil())
			defer r.Body.Close()

			m := map[string]map[string]gogadgets.Value{}
			dec := json.NewDecoder(r.Body)
			err = dec.Decode(&m)
			Expect(err).To(BeNil())
			Expect(len(m)).To(Equal(2))
			v, ok := m["back yard"]["sprinklers"]
			Expect(ok).To(BeTrue())
			Expect(v.Value).To(BeFalse())
		})

		It("lets you send a command to a gadget", func() {
			var buf bytes.Buffer
			enc := json.NewEncoder(&buf)
			m := map[string]string{
				"command": "turn on back yard sprinklers",
			}
			enc.Encode(m)

			req, err := http.NewRequest("POST", fmt.Sprintf(addr, "gadgets/sprinklers"), &buf)
			Expect(err).To(BeNil())
			req.AddCookie(cookies[0])
			r, err := http.DefaultClient.Do(req)
			Expect(err).To(BeNil())
			defer r.Body.Close()

			Expect(len(msgs)).To(Equal(1))
			msg := msgs[0]
			Expect(msg.Body).To(Equal("turn on back yard sprinklers"))
		})

		It("allows the sending of messages with a websocket", func() {
			u := strings.Replace(fmt.Sprintf(addr, "gadgets/sprinklers/websocket"), "http", "ws", -1)
			c := cookies[0]
			h := http.Header{"Origin": {u}, "Cookie": {c.String()}}
			ws, _, err := cstDialer.Dial(u, h)
			Expect(err).To(BeNil())
			defer ws.Close()

			uuid := gogadgets.GetUUID()
			msg := gogadgets.Message{
				Type:   gogadgets.COMMAND,
				Body:   "turn on back yard sprinklers",
				UUID:   uuid,
				Sender: "cli",
			}
			d, _ := json.Marshal(msg)
			err = ws.WriteMessage(websocket.TextMessage, d)
			Expect(err).To(BeNil())

			Eventually(func() int {
				return len(msgs)
			}).Should(Equal(1))

			msg = msgs[0]
			Expect(msg.Body).To(Equal("turn on back yard sprinklers"))
			Expect(msg.UUID).To(Equal(uuid))
		})

		It("does not allow the sending of messages with a websocket when there are no auth cookies", func() {
			u := strings.Replace(fmt.Sprintf(addr, "gadgets/sprinklers/websocket"), "http", "ws", -1)
			h := http.Header{"Origin": {u}}
			ws, r, err := cstDialer.Dial(u, h)
			Expect(r.StatusCode).To(Equal(http.StatusUnauthorized))
			Expect(err).ToNot(BeNil())
			Expect(ws).To(BeNil())
		})
	})
})
