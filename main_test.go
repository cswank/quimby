package main

import (
	"bytes"
	"encoding/base64"
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

var dialer = websocket.Dialer{
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
		port       string
		port2      string
		root       string
		iRoot      string
		dir        string
		pth        string
		u          *models.User
		u2         *models.User
		addr       string
		addr2      string
		db         *bolt.DB
		token      string
		readToken  string
		lights     *models.Gadget
		sprinklers *models.Gadget
		ts         *httptest.Server
		msgs       []gogadgets.Message
		clients    []map[string]string
		lg         *fakeLogger
	)

	BeforeEach(func() {
		lg = &fakeLogger{}
		msgs = []gogadgets.Message{}
		clients = []map[string]string{}
		ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" {
				if r.URL.Path == "/gadgets/locations/back yard/devices/sprinklers/status" {
					fmt.Fprintln(
						w,
						`{"value": true, "io": true}`,
					)
				} else {
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
				}
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
		addr = fmt.Sprintf("http://localhost:%s/api/%%s%%s%%s", port)
		addr2 = fmt.Sprintf("http://localhost:%s/%%s%%s%%s", port2)

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
			url := fmt.Sprintf(addr, "login", "", "")
			var err error
			r, err = http.Post(url, "application/json", &buf)
			return err
		}).Should(BeNil())
		Expect(r.StatusCode).To(Equal(http.StatusOK))
		token = r.Header.Get("Authorization")

		buf := bytes.Buffer{}
		enc := json.NewEncoder(&buf)
		usr2 := &models.User{
			Username: "him",
			Password: "shhhhhhhh",
		}
		enc.Encode(usr2)
		url := fmt.Sprintf(addr, "login", "", "")
		r, err = http.Post(url, "application/json", &buf)
		Expect(err).To(BeNil())

		Expect(r.StatusCode).To(Equal(http.StatusOK))
		readToken = r.Header.Get("Authorization")
	})

	AfterEach(func() {
		db.Close()
		os.RemoveAll(dir)
	})

	Context("logging in and out", func() {
		It("lets you log in", func() {
			//Already logged in above in BeforeEach
			Expect(len(token)).ToNot(Equal(0))
		})

		It("lets you log out", func() {
			req, err := http.NewRequest("POST", fmt.Sprintf(addr, "logout", "", ""), nil)
			Expect(err).To(BeNil())
			req.Header.Add("Authorization", token)
			r, err := http.DefaultClient.Do(req)
			Expect(err).To(BeNil())
			defer r.Body.Close()
			Expect(r.StatusCode).To(Equal(http.StatusOK))
		})
	})

	Context("logged in", func() {
		It("lets you get gadgets", func() {
			req, err := http.NewRequest("GET", fmt.Sprintf(addr, "gadgets", "", ""), nil)
			Expect(err).To(BeNil())
			req.Header.Add("Authorization", token)
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
			u := fmt.Sprintf(addr, "gadgets/", sprinklers.Id, "")
			req, err := http.NewRequest("GET", u, nil)
			Expect(err).To(BeNil())
			req.Header.Add("Authorization", token)
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

		It("gives a 404 for a non-gadget", func() {
			req, err := http.NewRequest("GET", fmt.Sprintf(addr, "gadgets/notarealid", "", ""), nil)
			Expect(err).To(BeNil())
			req.Header.Add("Authorization", token)
			r, err := http.DefaultClient.Do(req)
			Expect(err).To(BeNil())
			Expect(r.StatusCode).To(Equal(http.StatusNotFound))
			r.Body.Close()
		})

		It("lets you delete a gadget", func() {
			req, err := http.NewRequest("DELETE", fmt.Sprintf(addr, "gadgets/", sprinklers.Id, ""), nil)
			Expect(err).To(BeNil())
			req.Header.Add("Authorization", token)
			r, err := http.DefaultClient.Do(req)
			Expect(err).To(BeNil())
			defer r.Body.Close()

			req, err = http.NewRequest("GET", fmt.Sprintf(addr, "gadgets", "", ""), nil)
			Expect(err).To(BeNil())
			req.Header.Add("Authorization", token)
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
			req, err := http.NewRequest("DELETE", fmt.Sprintf(addr, "gadgets/", sprinklers.Id, ""), nil)
			Expect(err).To(BeNil())
			req.Header.Add("Authorization", readToken)
			r, err := http.DefaultClient.Do(req)
			Expect(err).To(BeNil())
			Expect(r.StatusCode).To(Equal(http.StatusUnauthorized))
			defer r.Body.Close()

			req, err = http.NewRequest("GET", fmt.Sprintf(addr, "gadgets", "", ""), nil)
			Expect(err).To(BeNil())
			req.Header.Add("Authorization", readToken)
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

			req, err := http.NewRequest("POST", fmt.Sprintf(addr, "gadgets", "", ""), &buf)
			Expect(err).To(BeNil())
			req.Header.Add("Authorization", token)
			r, err := http.DefaultClient.Do(req)
			Expect(err).To(BeNil())
			defer r.Body.Close()
			Expect(r.StatusCode).To(Equal(http.StatusOK))
			Expect(len(r.Header["Location"])).To(Equal(1))
			u := r.Header["Location"][0]
			Expect(u).To(Equal("/api/gadgets/back%20yard"))

			req, err = http.NewRequest("GET", fmt.Sprintf(addr, "gadgets", "", ""), nil)
			Expect(err).To(BeNil())
			req.Header.Add("Authorization", token)
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
			u := fmt.Sprintf(addr, "gadgets/", sprinklers.Id, "/status")
			req, err := http.NewRequest("GET", u, nil)
			Expect(err).To(BeNil())
			req.Header.Add("Authorization", readToken)
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

			req, err := http.NewRequest("POST", fmt.Sprintf(addr, "gadgets/", sprinklers.Id, ""), &buf)
			Expect(err).To(BeNil())
			req.Header.Add("Authorization", token)
			r, err := http.DefaultClient.Do(req)
			Expect(err).To(BeNil())
			defer r.Body.Close()

			Eventually(func() int {
				return len(msgs)
			}).Should(Equal(1))

			msg := msgs[0]
			Expect(msg.Body).To(Equal("turn on back yard sprinklers"))
		})

		It("lets you get the value of a device", func() {
			u := fmt.Sprintf(
				addr,
				"gadgets/",
				sprinklers.Id,
				"/locations/back%20yard/devices/sprinklers/status",
			)
			req, err := http.NewRequest("GET", u, nil)
			Expect(err).To(BeNil())
			req.Header.Add("Authorization", token)
			r, err := http.DefaultClient.Do(req)
			Expect(err).To(BeNil())
			defer r.Body.Close()

			var v gogadgets.Value
			dec := json.NewDecoder(r.Body)
			err = dec.Decode(&v)
			Expect(err).To(BeNil())
			Expect(v.Value).To(BeTrue())
		})

		It("lets you turn on a device", func() {
			var buf bytes.Buffer
			enc := json.NewEncoder(&buf)
			v := gogadgets.Value{
				Value: true,
			}
			enc.Encode(v)

			u := fmt.Sprintf(
				addr,
				"gadgets/",
				sprinklers.Id,
				"/locations/front%20yard/devices/sprinklers/status",
			)
			req, err := http.NewRequest("POST", u, &buf)
			Expect(err).To(BeNil())
			req.Header.Add("Authorization", token)
			r, err := http.DefaultClient.Do(req)
			Expect(err).To(BeNil())
			defer r.Body.Close()

			Expect(len(msgs)).To(Equal(1))
			msg := msgs[0]
			Expect(msg.Body).To(Equal("turn on front yard sprinklers"))
		})

		It("lets you turn off a device", func() {
			var buf bytes.Buffer
			enc := json.NewEncoder(&buf)
			v := gogadgets.Value{
				Value: false,
			}
			enc.Encode(v)

			u := fmt.Sprintf(
				addr,
				"gadgets/",
				sprinklers.Id,
				"/locations/front%20yard/devices/sprinklers/status",
			)
			req, err := http.NewRequest("POST", u, &buf)
			Expect(err).To(BeNil())
			req.Header.Add("Authorization", token)
			r, err := http.DefaultClient.Do(req)
			Expect(err).To(BeNil())
			defer r.Body.Close()

			Expect(len(msgs)).To(Equal(1))
			msg := msgs[0]
			Expect(msg.Body).To(Equal("turn off front yard sprinklers"))
		})

		It("allows the sending of a message with a websocket", func() {
			//get the ticket for a websocket
			u := fmt.Sprintf(
				addr,
				"gadgets/",
				sprinklers.Id,
				"/websocket",
			)

			req, err := http.NewRequest("GET", u, nil)
			Expect(err).To(BeNil())
			req.Header.Add("Authorization", token)
			r, err := http.DefaultClient.Do(req)
			Expect(err).To(BeNil())
			r.Body.Close()
			l := r.Header.Get("Location")

			a := strings.Replace(addr, "http", "ws", -1)
			a = strings.Replace(a, "/api/", "", -1)
			u = fmt.Sprintf(
				a,
				l,
				"",
				"",
			)
			h := http.Header{"Origin": {u}}
			ws, _, err := dialer.Dial(u, h)
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

		It("allows the getting of a message with a websocket", func() {
			//get the ticket for a websocket
			u := fmt.Sprintf(
				addr,
				"gadgets/",
				sprinklers.Id,
				"/websocket",
			)

			req, err := http.NewRequest("GET", u, nil)
			Expect(err).To(BeNil())
			req.Header.Add("Authorization", token)
			r, err := http.DefaultClient.Do(req)
			Expect(err).To(BeNil())
			r.Body.Close()
			l := r.Header.Get("Location")

			a := strings.Replace(addr, "http", "ws", -1)
			a = strings.Replace(a, "/api/", "", -1)
			u = fmt.Sprintf(
				a,
				l,
				"",
				"",
			)

			h := http.Header{"Origin": {u}}
			ws, _, err := dialer.Dial(u, h)
			Expect(err).To(BeNil())
			defer ws.Close()

			uuid := gogadgets.GetUUID()
			msg := gogadgets.Message{
				Location: "back yard",
				Name:     "sprinklers",
				Type:     gogadgets.UPDATE,
				Host:     ts.URL,
				Value: gogadgets.Value{
					Value: true,
				},
				UUID:   uuid,
				Sender: "sprinklers",
			}
			var buf bytes.Buffer
			enc := json.NewEncoder(&buf)
			enc.Encode(msg)
			req, err = http.NewRequest("POST", fmt.Sprintf(addr2, "internal/updates", "", ""), &buf)
			Expect(err).To(BeNil())
			req.Header.Add("Authorization", token)
			r, err = http.DefaultClient.Do(req)
			Expect(err).To(BeNil())
			Expect(r.StatusCode).To(Equal(http.StatusOK))
			defer r.Body.Close()

			var msg2 gogadgets.Message
			err = ws.ReadJSON(&msg2)
			Expect(msg2.Value.Value).To(BeTrue())
			Expect(msg2.Location).To(Equal("back yard"))
			Expect(msg2.UUID).To(Equal(uuid))
		})
	})

	Context("hackers hacking in", func() {
		Describe("does not let you", func() {
			It("get gadgets when you mess with the token", func() {
				req, err := http.NewRequest("GET", fmt.Sprintf(addr, "gadgets", "", ""), nil)
				Expect(err).To(BeNil())
				l := len(token)
				req.Header.Add("Authorization", token[:l-2])
				r, err := http.DefaultClient.Do(req)
				Expect(err).To(BeNil())
				defer r.Body.Close()
				Expect(r.StatusCode).To(Equal(http.StatusUnauthorized))
				d, _ := ioutil.ReadAll(r.Body)
				Expect(string(d)).To(Equal("Not Authorized"))
			})
			It("get gadgets when you change the token", func() {
				req, err := http.NewRequest("GET", fmt.Sprintf(addr, "gadgets", "", ""), nil)
				Expect(err).To(BeNil())
				parts := strings.Split(readToken, ".")
				p := []byte(`{"exp":2443503371,"iat":1443416971,"sub":"me"}`)
				s := base64.StdEncoding.EncodeToString(p)
				parts[1] = strings.Replace(s, "=", "", -1)
				t := strings.Join(parts, ".")
				req.Header.Add("Authorization", t)
				r, err := http.DefaultClient.Do(req)
				Expect(err).To(BeNil())
				defer r.Body.Close()
				Expect(r.StatusCode).To(Equal(http.StatusUnauthorized))
				d, _ := ioutil.ReadAll(r.Body)
				Expect(string(d)).To(Equal("Not Authorized"))
			})
		})
	})

	Context("not logged in", func() {
		Describe("does not let you", func() {
			It("get gadgets", func() {
				r, err := http.Get(fmt.Sprintf(addr, "gadgets", "", ""))
				Expect(err).To(BeNil())
				Expect(r.StatusCode).To(Equal(http.StatusUnauthorized))
			})
			It(" get a gadget", func() {
				r, err := http.Get(fmt.Sprintf(addr, "gadgets/", sprinklers.Id, ""))
				Expect(err).To(BeNil())
				Expect(r.StatusCode).To(Equal(http.StatusUnauthorized))
				d, _ := ioutil.ReadAll(r.Body)
				Expect(strings.TrimSpace(string(d))).To(Equal("Not Authorized"))
				r.Body.Close()
			})

			It(" delete a gadget", func() {
				req, err := http.NewRequest("DELETE", fmt.Sprintf(addr, "gadgets/", sprinklers.Id, ""), nil)
				Expect(err).To(BeNil())
				r, err := http.DefaultClient.Do(req)
				Expect(err).To(BeNil())
				Expect(r.StatusCode).To(Equal(http.StatusUnauthorized))
				d, _ := ioutil.ReadAll(r.Body)
				Expect(strings.TrimSpace(string(d))).To(Equal("Not Authorized"))
				defer r.Body.Close()
			})

			It(" add a gadget", func() {
				var buf bytes.Buffer
				enc := json.NewEncoder(&buf)
				g := models.Gadget{
					Name: "back yard",
					Host: "http://overthere.com",
				}
				enc.Encode(g)
				r, err := http.Post(fmt.Sprintf(addr, "gadgets", "", ""), "application/json", &buf)
				Expect(r.StatusCode).To(Equal(http.StatusUnauthorized))
				d, _ := ioutil.ReadAll(r.Body)
				Expect(strings.TrimSpace(string(d))).To(Equal("Not Authorized"))
				Expect(err).To(BeNil())
				r.Body.Close()
			})

			It(" get the status of a gadget", func() {
				u := fmt.Sprintf(
					addr,
					"gadgets/",
					sprinklers.Id,
					"/status",
				)
				r, err := http.Get(u)
				Expect(err).To(BeNil())
				Expect(r.StatusCode).To(Equal(http.StatusUnauthorized))
				d, _ := ioutil.ReadAll(r.Body)
				Expect(strings.TrimSpace(string(d))).To(Equal("Not Authorized"))
				r.Body.Close()
			})

			It(" send a command to a gadget", func() {
				var buf bytes.Buffer
				enc := json.NewEncoder(&buf)
				m := map[string]string{
					"command": "turn on back yard sprinklers",
				}
				enc.Encode(m)
				u := fmt.Sprintf(addr, "gadgets/", sprinklers.Id, "")
				r, err := http.Post(u, "application/json", &buf)
				Expect(r.StatusCode).To(Equal(http.StatusUnauthorized))
				Expect(err).To(BeNil())
				d, _ := ioutil.ReadAll(r.Body)
				Expect(strings.TrimSpace(string(d))).To(Equal("Not Authorized"))
				r.Body.Close()
			})

			It("does not allow the getting of a websocket", func() {
				u := strings.Replace(
					fmt.Sprintf(
						addr,
						"gadgets/",
						sprinklers.Id,
						"/websocket",
					), "http", "ws", -1)
				h := http.Header{"Origin": {u}}
				ws, r, err := dialer.Dial(u, h)
				Expect(r.StatusCode).To(Equal(http.StatusUnauthorized))
				Expect(err).ToNot(BeNil())
				Expect(ws).To(BeNil())
			})

			It(" turn on a device", func() {
				var buf bytes.Buffer
				enc := json.NewEncoder(&buf)
				v := gogadgets.Value{
					Value: true,
				}
				enc.Encode(v)

				u := fmt.Sprintf(
					addr,
					"gadgets/",
					sprinklers.Id,
					"/locations/front%20yard/devices/sprinklers/status",
				)
				r, err := http.Post(u, "application/json", &buf)
				Expect(err).To(BeNil())
				d, _ := ioutil.ReadAll(r.Body)
				Expect(strings.TrimSpace(string(d))).To(Equal("Not Authorized"))
				r.Body.Close()
				Expect(len(msgs)).To(Equal(0))
			})
		})
	})
})
