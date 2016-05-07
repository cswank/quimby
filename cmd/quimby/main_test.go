package main

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/cswank/gogadgets"
	"github.com/cswank/quimby"
	"github.com/gorilla/websocket"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	p int
)

func init() {
	p = 1024 + rand.Intn(65535-2024)
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

type requester func() *http.Response
type stringGetter func() string
type bufGetter func() io.Reader
type floatGetter func() float64
type cookieGetter func() *http.Cookie

var _ = Describe("Quimby", func() {
	var (
		getAccept  stringGetter
		getReq     requester
		getURL     stringGetter
		getBuf     bufGetter
		getTok     stringGetter
		getMethod  stringGetter
		port       string
		port2      string
		root       string
		iRoot      string
		dir        string
		pth        string
		u          *quimby.User
		u2         *quimby.User
		u3         *quimby.User
		addr       string
		addr2      string
		beerAddr   string
		adminAddr  string
		db         *bolt.DB
		token      string
		readToken  string
		adminToken string
		lights     *quimby.Gadget
		sprinklers *quimby.Gadget
		ts         *httptest.Server
		msgs       []gogadgets.Message
		clients    []map[string]string
		lg         *fakeLogger
	)

	BeforeEach(func() {
		getAccept = func() string { return "application/json" }
		port = fmt.Sprintf("%d", p)
		port2 = fmt.Sprintf("%d", p+1)
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

		root = fmt.Sprintf("%s", RandString(10))
		iRoot = fmt.Sprintf("%s", RandString(10))

		os.Setenv("QUIMBY_PORT", port)
		os.Setenv("QUIMBY_INTERNAL_PORT", port2)
		os.Setenv("QUIMBY_HOST", "http://localhost")
		addr = fmt.Sprintf("http://localhost:%s/api/%%s%%s%%s", port)
		addr2 = fmt.Sprintf("http://localhost:%s/%%s%%s%%s", port2)
		beerAddr = fmt.Sprintf("http://localhost:%s/api/beer/%%s", port)
		adminAddr = fmt.Sprintf("http://localhost:%s/api/admin/%%s", port)

		dir, _ = ioutil.TempDir("", "")
		pth = path.Join(dir, "db")

		var err error
		db, err = quimby.GetDB(pth)
		Expect(err).To(BeNil())

		u = &quimby.User{
			Username:   "me",
			Password:   "hushhush",
			Permission: "write",
			DB:         db,
		}

		err = u.Save()
		Expect(err).To(BeNil())

		u2 = &quimby.User{
			Username:   "him",
			Password:   "shhhhhhhh",
			Permission: "read",
			DB:         db,
		}

		err = u2.Save()
		Expect(err).To(BeNil())

		u3 = &quimby.User{
			Username:   "boss",
			Password:   "sosecret",
			Permission: "admin",
			DB:         db,
		}

		err = u3.Save()
		Expect(err).To(BeNil())

		lights = &quimby.Gadget{
			Name: "lights",
			Host: ts.URL,
			DB:   db,
		}
		err = lights.Save()
		Expect(err).To(BeNil())

		sprinklers = &quimby.Gadget{
			Name: "sprinklers",
			Host: ts.URL,
			DB:   db,
		}
		err = sprinklers.Save()
		Expect(err).To(BeNil())
		clients := quimby.NewClientHolder()
		go start(db, port, port2, root, iRoot, lg, clients)
		Eventually(func() error {
			url := fmt.Sprintf(addr, "ping", "", "")
			_, err := http.Get(url)
			return err
		}).Should(BeNil())

		getMethod = func() string {
			return "GET"
		}

		getReq = func() *http.Response {
			req, err := http.NewRequest(getMethod(), getURL(), getBuf())
			Expect(err).To(BeNil())
			req.Header.Add("Authorization", getTok())
			req.Header.Add("Accept", getAccept())
			r, err := http.DefaultClient.Do(req)
			Expect(err).To(BeNil())
			return r
		}

	})

	AfterEach(func() {
		p += 2
		db.Close()
		os.RemoveAll(dir)
	})

	Describe("with jwt", func() {
		BeforeEach(func() {
			url := fmt.Sprintf(addr, "login?auth=jwt", "", "")
			buf := bytes.Buffer{}
			enc := json.NewEncoder(&buf)
			usr := &quimby.User{
				Username: "me",
				Password: "hushhush",
			}
			enc.Encode(usr)

			r, err := http.Post(url, "application/json", &buf)
			Expect(err).To(BeNil())

			Expect(r.StatusCode).To(Equal(http.StatusOK))
			token = r.Header.Get("Authorization")

			buf = bytes.Buffer{}
			enc = json.NewEncoder(&buf)
			usr2 := &quimby.User{
				Username: "him",
				Password: "shhhhhhhh",
			}
			enc.Encode(usr2)

			r, err = http.Post(url, "application/json", &buf)
			Expect(err).To(BeNil())
			Expect(r.StatusCode).To(Equal(http.StatusOK))
			readToken = r.Header.Get("Authorization")

			buf = bytes.Buffer{}
			enc = json.NewEncoder(&buf)
			usr3 := &quimby.User{
				Username: "boss",
				Password: "sosecret",
			}
			enc.Encode(usr3)

			r, err = http.Post(url, "application/json", &buf)
			Expect(err).To(BeNil())
			Expect(r.StatusCode).To(Equal(http.StatusOK))
			adminToken = r.Header.Get("Authorization")

			getURL = func() string {
				return fmt.Sprintf(adminAddr, "clients")
			}

			getBuf = func() io.Reader {
				return nil
			}

			getTok = func() string { return token }
		})

		Context("hackers hacking in", func() {
			BeforeEach(func() {
				getURL = func() string { return fmt.Sprintf(addr, "gadgets", "", "") }
			})

			Describe("does not let you", func() {
				It("get gadgets when you mess with the token", func() {

					l := len(token)
					getTok = func() string { return token[:l-2] }
					r := getReq()
					defer r.Body.Close()
					Expect(r.StatusCode).To(Equal(http.StatusUnauthorized))
					d, _ := ioutil.ReadAll(r.Body)
					Expect(string(d)).To(Equal("Not Authorized"))
				})

				It("get gadgets when you change the token", func() {

					getTok = func() string {
						parts := strings.Split(readToken, ".")
						p := []byte(`{"exp":2443503371,"iat":1443416971,"sub":"me"}`)
						s := base64.StdEncoding.EncodeToString(p)
						parts[1] = strings.Replace(s, "=", "", -1)
						return strings.Join(parts, ".")
					}
					r := getReq()
					defer r.Body.Close()
					Expect(r.StatusCode).To(Equal(http.StatusUnauthorized))
					d, _ := ioutil.ReadAll(r.Body)
					Expect(string(d)).To(Equal("Not Authorized"))
				})
			})
		})

		Context("admin stuff", func() {
			BeforeEach(func() {
				getTok = func() string { return adminToken }
			})
			Describe("lets admins", func() {
				It("see how many websocket clients there are", func() {
					r := getReq()
					defer r.Body.Close()
					Expect(r.StatusCode).To(Equal(http.StatusOK))
					var m map[string]int
					dec := json.NewDecoder(r.Body)
					Expect(dec.Decode(&m)).To(BeNil())
					Expect(len(m)).To(Equal(0))
				})

				It("get a list of users", func() {
					getURL = func() string { return fmt.Sprintf(addr, "users", "", "") }
					r := getReq()
					defer r.Body.Close()
					Expect(r.StatusCode).To(Equal(http.StatusOK))

					var u []quimby.User
					dec := json.NewDecoder(r.Body)
					Expect(dec.Decode(&u)).To(BeNil())
					Expect(len(u)).To(Equal(3))
				})

				It("get a user", func() {
					getURL = func() string { return fmt.Sprintf(addr, "users/", "boss", "") }
					r := getReq()
					defer r.Body.Close()
					Expect(r.StatusCode).To(Equal(http.StatusOK))

					var u quimby.User
					dec := json.NewDecoder(r.Body)
					Expect(dec.Decode(&u)).To(BeNil())
					Expect(u.Username).To(Equal("boss"))
				})
			})

			Describe("does not let non-admins", func() {

				BeforeEach(func() {
					getTok = func() string { return token }
				})

				It("get a list of users", func() {
					getURL = func() string { return fmt.Sprintf(addr, "users", "", "") }
					r := getReq()
					defer r.Body.Close()
					Expect(r.StatusCode).To(Equal(http.StatusUnauthorized))

					var u []quimby.User
					dec := json.NewDecoder(r.Body)
					Expect(dec.Decode(&u)).ToNot(BeNil())
					Expect(len(u)).To(Equal(0))
				})

				It("get a user", func() {
					getURL = func() string { return fmt.Sprintf(addr, "users/", "boss", "") }
					r := getReq()
					defer r.Body.Close()
					Expect(r.StatusCode).To(Equal(http.StatusUnauthorized))

					var u quimby.User
					dec := json.NewDecoder(r.Body)
					Expect(dec.Decode(&u)).ToNot(BeNil())
					Expect(u.Username).To(Equal(""))
				})

				It("see how many websocket clients there are", func() {
					r := getReq()
					Expect(r.StatusCode).To(Equal(http.StatusUnauthorized))
				})
			})
		})

		Context("logging in and out", func() {
			It("lets you log in", func() {
				//Already logged in above in BeforeEach
				Expect(len(token)).ToNot(Equal(0))
			})

			It("lets you log out", func() {
				getURL = func() string { return fmt.Sprintf(addr, "logout", "", "") }
				getTok = func() string { return token }
				getMethod = func() string { return "POST" }
				r := getReq()
				defer r.Body.Close()
				Expect(r.StatusCode).To(Equal(http.StatusOK))
			})
		})

		Context("OPTIONS", func() {
			// FIt("gets them for gadgets", func() {
			// 				req, err := http.NewRequest("OPTIONS", fmt.Sprintf(addr, "gadgets/", sprinklers.Id, ""), nil)
			// 				Expect(err).To(BeNil())
			// 				req.Header.Add("Authorization", token)
			// 				r, err := http.DefaultClient.Do(req)
			// 				Expect(err).To(BeNil())
			// 				defer r.Body.Close()
			// 				d, _ := ioutil.ReadAll(r.Body)
			// 				expected := `{
			//   "POST": {
			//     "description": "send a command to the gogadgets target",
			//     "body": "{\"command\": \"turn on kitchen light\"}",
			//     "response": "no body"
			//   },
			//   "GET": {
			//     "description": "get the metadata for a gogadgets target",
			//     "response": "{\"host\": \"http://127.0.0.1:55507\", \"id\": \"f174b39e-4d7d-4d7c-9fbf-ba0947058498\", \"name\": \"sprinklers\"}"
			//   }
			// }`
			// 				Expect(string(d)).To(MatchJSON(expected))
			// 			})
		})

		Context("logged in", func() {
			It("lets you get gadgets", func() {
				getURL = func() string { return fmt.Sprintf(addr, "gadgets", "", "") }
				r := getReq()
				defer r.Body.Close()
				Expect(r.StatusCode).To(Equal(http.StatusOK))
				var g []quimby.Gadget
				dec := json.NewDecoder(r.Body)
				Expect(dec.Decode(&g)).To(BeNil())
				Expect(len(g)).To(Equal(2))
			})

			It("lets you get a gadget", func() {
				getURL = func() string { return fmt.Sprintf(addr, "gadgets/", sprinklers.Id, "") }
				r := getReq()
				defer r.Body.Close()

				var g quimby.Gadget
				dec := json.NewDecoder(r.Body)
				Expect(dec.Decode(&g)).To(BeNil())
				Expect(g.Name).To(Equal("sprinklers"))
				Expect(g.Host).To(Equal(ts.URL))
			})

			It("lets you add notes to a gadget", func() {
				getURL = func() string { return fmt.Sprintf(addr, "gadgets/", sprinklers.Id, "/notes") }
				getBuf = func() io.Reader {
					var buf bytes.Buffer
					enc := json.NewEncoder(&buf)
					note := map[string]string{
						"text": "jibber jabber",
					}
					enc.Encode(note)
					return &buf
				}
				getMethod = func() string { return "POST" }
				r := getReq()
				Expect(r.StatusCode).To(Equal(http.StatusOK))
				r.Body.Close()
			})

			It("gives a 404 for a non-gadget", func() {
				getURL = func() string { return fmt.Sprintf(addr, "gadgets/notarealid", "", "") }
				r := getReq()
				Expect(r.StatusCode).To(Equal(http.StatusNotFound))
				r.Body.Close()
			})

			Context("deleting", func() {

				var (
					numGadgets int
				)

				BeforeEach(func() {
					getMethod = func() string { return "DELETE" }
					getURL = func() string { return fmt.Sprintf(addr, "gadgets/", sprinklers.Id, "") }
				})

				AfterEach(func() {
					getMethod = func() string { return "GET" }
					getURL = func() string { return fmt.Sprintf(addr, "gadgets", "", "") }
					r := getReq()
					defer r.Body.Close()
					Expect(r.StatusCode).To(Equal(http.StatusOK))
					var gadgs []quimby.Gadget
					dec := json.NewDecoder(r.Body)
					Expect(dec.Decode(&gadgs)).To(BeNil())
					Expect(len(gadgs)).To(Equal(numGadgets))
				})

				It("lets you delete a gadget", func() {
					r := getReq()
					defer r.Body.Close()
					numGadgets = 1
				})

				It("doesn't let a read only user delete a gadget", func() {
					getTok = func() string { return readToken }
					getMethod = func() string { return "DELETE" }
					getURL = func() string { return fmt.Sprintf(addr, "gadgets/", sprinklers.Id, "") }
					r := getReq()
					r.Body.Close()
					Expect(r.StatusCode).To(Equal(http.StatusUnauthorized))
					numGadgets = 2
				})
			})

			Context("adding and updating", func() {
				BeforeEach(func() {
					getBuf = func() io.Reader {
						var buf bytes.Buffer
						enc := json.NewEncoder(&buf)
						g := quimby.Gadget{
							Name: "back yard",
							Host: "http://overthere.com",
						}
						enc.Encode(g)
						return &buf
					}
					getMethod = func() string { return "POST" }
					getURL = func() string { return fmt.Sprintf(addr, "gadgets", "", "") }
				})

				It("lets you add a gadget", func() {
					r := getReq()
					defer r.Body.Close()
					Expect(r.StatusCode).To(Equal(http.StatusOK))
					Expect(len(r.Header["Location"])).To(Equal(1))
					u := r.Header["Location"][0]
					parts := strings.Split(u, "/")
					Expect(len(parts)).To(Equal(4))
					id := parts[3]

					getMethod = func() string { return "GET" }
					getBuf = func() io.Reader { return nil }
					getURL = func() string { return fmt.Sprintf(addr, "gadgets/", id, "") }

					r = getReq()
					defer r.Body.Close()
					Expect(r.StatusCode).To(Equal(http.StatusOK))
					var g2 quimby.Gadget
					dec := json.NewDecoder(r.Body)
					Expect(dec.Decode(&g2)).To(BeNil())
					Expect(g2.Name).To(Equal("back yard"))
				})

				It("lets you update a gadget", func() {

					r := getReq()
					defer r.Body.Close()
					Expect(r.StatusCode).To(Equal(http.StatusOK))
					Expect(len(r.Header["Location"])).To(Equal(1))
					u := r.Header["Location"][0]
					parts := strings.Split(u, "/")
					Expect(len(parts)).To(Equal(4))
					id := parts[3]

					getBuf = func() io.Reader {
						var buf bytes.Buffer
						enc := json.NewEncoder(&buf)
						g := quimby.Gadget{
							Id:   id,
							Name: "back yard for reals",
							Host: "http://overthere.com",
						}
						enc.Encode(g)
						return &buf
					}

					getURL = func() string { return fmt.Sprintf(addr, "gadgets/", id, "") }
					r = getReq()
					defer r.Body.Close()
					Expect(r.StatusCode).To(Equal(http.StatusOK))
					Expect(len(r.Header["Location"])).To(Equal(1))
					u = r.Header["Location"][0]
					Expect(u).To(Equal("/api/gadgets/" + id))

					getURL = func() string { return fmt.Sprintf(addr, "gadgets/", id, "") }
					getMethod = func() string { return "GET" }
					getBuf = func() io.Reader { return nil }
					r = getReq()
					defer r.Body.Close()
					Expect(r.StatusCode).To(Equal(http.StatusOK))
					var g2 quimby.Gadget
					dec := json.NewDecoder(r.Body)
					Expect(dec.Decode(&g2)).To(BeNil())
					Expect(g2.Name).To(Equal("back yard for reals"))
				})
			})

			It("lets you get the status of a gadget", func() {
				getURL = func() string { return fmt.Sprintf(addr, "gadgets/", sprinklers.Id, "/status") }
				getTok = func() string { return readToken }
				r := getReq()
				defer r.Body.Close()

				m := map[string]map[string]gogadgets.Value{}
				dec := json.NewDecoder(r.Body)
				Expect(dec.Decode(&m)).To(BeNil())
				Expect(len(m)).To(Equal(2))
				v, ok := m["back yard"]["sprinklers"]
				Expect(ok).To(BeTrue())
				Expect(v.Value).To(BeFalse())
			})

			It("lets you send a command to a gadget", func() {
				getBuf = func() io.Reader {
					var buf bytes.Buffer
					enc := json.NewEncoder(&buf)
					m := map[string]string{
						"command": "turn on back yard sprinklers",
					}
					enc.Encode(m)
					return &buf
				}

				getMethod = func() string { return "POST" }
				getURL = func() string { return fmt.Sprintf(addr, "gadgets/", sprinklers.Id, "/command") }
				r := getReq()
				defer r.Body.Close()

				Eventually(func() int {
					return len(msgs)
				}).Should(Equal(1))

				msg := msgs[0]
				Expect(msg.Body).To(Equal("turn on back yard sprinklers"))
			})

			It("lets you get the value of a device", func() {
				getURL = func() string {
					return fmt.Sprintf(
						addr,
						"gadgets/",
						sprinklers.Id,
						"/locations/back%20yard/devices/sprinklers/status",
					)
				}
				r := getReq()
				defer r.Body.Close()

				var v gogadgets.Value
				dec := json.NewDecoder(r.Body)
				Expect(dec.Decode(&v)).To(BeNil())
				Expect(v.Value).To(BeTrue())
			})

			Context("on and off", func() {
				var (
					val bool
				)

				BeforeEach(func() {
					val = false
					getBuf = func() io.Reader {
						var buf bytes.Buffer
						enc := json.NewEncoder(&buf)
						v := gogadgets.Value{
							Value: val,
						}
						enc.Encode(v)
						return &buf
					}

					getURL = func() string {
						return fmt.Sprintf(
							addr,
							"gadgets/",
							sprinklers.Id,
							"/locations/front%20yard/devices/sprinklers/status",
						)
					}
					getMethod = func() string { return "POST" }
				})

				It("lets you turn off a device", func() {
					val = false
					r := getReq()
					defer r.Body.Close()

					Expect(len(msgs)).To(Equal(1))
					msg := msgs[0]
					Expect(msg.Body).To(Equal("turn off front yard sprinklers"))
				})

				It("lets you turn on a device", func() {
					val = true
					r := getReq()
					defer r.Body.Close()

					Expect(len(msgs)).To(Equal(1))
					msg := msgs[0]
					Expect(msg.Body).To(Equal("turn on front yard sprinklers"))
				})
			})

			It("saves and gets stats", func() {
				var getVal floatGetter

				getBuf = func() io.Reader {
					v := map[string]float64{"value": getVal()}
					var buf bytes.Buffer
					enc := json.NewEncoder(&buf)
					enc.Encode(v)
					return &buf
				}

				getVal = func() float64 { return 33.3 }
				getURL = func() string {
					return fmt.Sprintf(
						addr2,
						"internal/gadgets/",
						sprinklers.Id,
						"/sources/front%20yard%20temperature",
					)
				}
				getMethod = func() string { return "POST" }
				n := time.Now()

				r := getReq()
				defer r.Body.Close()
				Expect(r.StatusCode).To(Equal(http.StatusOK))
				time.Sleep(10 * time.Millisecond)

				getVal = func() float64 { return 33.5 }
				r = getReq()
				defer r.Body.Close()

				getURL = func() string {
					return fmt.Sprintf(
						addr,
						"gadgets/",
						sprinklers.Id,
						fmt.Sprintf(
							"/sources/front%%20yard%%20temperature?start=%s&end=%s",
							url.QueryEscape(n.Add(-60*time.Second).Format(time.RFC3339Nano)),
							url.QueryEscape(n.Add(time.Second).Format(time.RFC3339Nano)),
						),
					)
				}

				getMethod = func() string { return "GET" }
				r = getReq()
				defer r.Body.Close()

				var points []quimby.DataPoint
				dec := json.NewDecoder(r.Body)
				Expect(dec.Decode(&points)).To(BeNil())
				Expect(len(points)).To(Equal(2))

				p1 := points[0]
				Expect(p1.Value).To(Equal(33.3))

				p2 := points[1]
				Expect(p2.Value).To(Equal(33.5))

				//Get them as csv
				getAccept = func() string { return "application/csv" }
				r = getReq()
				defer r.Body.Close()
				c := csv.NewReader(r.Body)
				rows, err := c.ReadAll()
				Expect(err).To(BeNil())
				Expect(len(rows)).To(Equal(3))

				h := rows[0]
				Expect(h[0]).To(Equal("time"))
				Expect(h[1]).To(Equal("value"))

				r1 := rows[1]
				Expect(r1[1]).To(Equal("33.300000"))

				r2 := rows[2]
				Expect(r2[1]).To(Equal("33.500000"))
			})

			Describe("websockets", func() {
				var (
					ws *websocket.Conn
				)

				BeforeEach(func() {
					a := strings.Replace(addr, "http", "ws", -1)
					u := fmt.Sprintf(
						a,
						"gadgets/",
						sprinklers.Id,
						"/websocket",
					)

					h := http.Header{"Origin": {u}, "Authorization": {token}}
					var err error
					ws, _, err = dialer.Dial(u, h)
					Expect(err).To(BeNil())
				})

				AfterEach(func() {
					ws.Close()
				})

				It("allows the sending of a message with a websocket", func() {

					uuid := gogadgets.GetUUID()
					msg := gogadgets.Message{
						Type:   gogadgets.COMMAND,
						Body:   "turn on back yard sprinklers",
						UUID:   uuid,
						Sender: "cli",
					}
					d, _ := json.Marshal(msg)
					err := ws.WriteMessage(websocket.TextMessage, d)
					Expect(err).To(BeNil())

					Eventually(func() int {
						return len(msgs)
					}).Should(Equal(1))

					msg = msgs[0]
					Expect(msg.Body).To(Equal("turn on back yard sprinklers"))
					Expect(msg.UUID).To(Equal(uuid))
				})

				It("allows the getting of a message with a websocket", func() {

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
					req, err := http.NewRequest("POST", fmt.Sprintf(addr2, "internal/updates", "", ""), &buf)
					Expect(err).To(BeNil())
					req.Header.Add("Authorization", token)
					r, err := http.DefaultClient.Do(req)
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

		})
	})

	Describe("with cookies", func() {

		var (
			cookies      []*http.Cookie
			readCookies  []*http.Cookie
			adminCookies []*http.Cookie
			getCookie    cookieGetter
		)

		BeforeEach(func() {
			url := fmt.Sprintf(addr, "login", "", "")
			buf := bytes.Buffer{}
			enc := json.NewEncoder(&buf)
			usr := &quimby.User{
				Username: "me",
				Password: "hushhush",
			}
			enc.Encode(usr)
			r, err := http.Post(url, "application/json", &buf)
			Expect(err).To(BeNil())
			Expect(r.StatusCode).To(Equal(http.StatusOK))
			cookies = r.Cookies()

			buf = bytes.Buffer{}
			enc = json.NewEncoder(&buf)
			usr2 := &quimby.User{
				Username: "him",
				Password: "shhhhhhhh",
			}
			enc.Encode(usr2)

			r, err = http.Post(url, "application/json", &buf)
			Expect(err).To(BeNil())

			Expect(r.StatusCode).To(Equal(http.StatusOK))
			readCookies = r.Cookies()

			buf = bytes.Buffer{}
			enc = json.NewEncoder(&buf)
			usr3 := &quimby.User{
				Username: "boss",
				Password: "sosecret",
			}
			enc.Encode(usr3)

			r, err = http.Post(url, "application/json", &buf)
			Expect(err).To(BeNil())

			Expect(r.StatusCode).To(Equal(http.StatusOK))
			adminCookies = r.Cookies()

			getBuf = func() io.Reader { return nil }
			getMethod = func() string { return "GET" }

			getReq = func() *http.Response {
				req, err := http.NewRequest(getMethod(), getURL(), getBuf())
				Expect(err).To(BeNil())
				req.AddCookie(getCookie())
				req.Header.Add("Accept", getAccept())
				r, err := http.DefaultClient.Do(req)
				Expect(err).To(BeNil())
				return r
			}
		})

		Context("admin stuff", func() {
			BeforeEach(func() {

				getURL = func() string {
					return fmt.Sprintf(adminAddr, "clients")
				}
			})

			It("lets you see how many websocket clients there are", func() {
				getCookie = func() *http.Cookie {
					return adminCookies[0]
				}
				r := getReq()
				defer r.Body.Close()
				Expect(r.StatusCode).To(Equal(http.StatusOK))
				var m map[string]int
				dec := json.NewDecoder(r.Body)
				Expect(dec.Decode(&m)).To(BeNil())
				Expect(len(m)).To(Equal(0))
			})

			It("does not let non-admins see how many websocket clients there are", func() {
				getCookie = func() *http.Cookie {
					return cookies[0]
				}

				r := getReq()
				r.Body.Close()
				Expect(r.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("logging in and out", func() {
			It("lets you log in", func() {
				//Already logged in above in BeforeEach
				Expect(len(token)).ToNot(Equal(0))
			})

			It("lets you log out", func() {
				req, err := http.NewRequest("POST", fmt.Sprintf(addr, "logout", "", ""), nil)
				Expect(err).To(BeNil())
				req.AddCookie(cookies[0])
				r, err := http.DefaultClient.Do(req)
				Expect(err).To(BeNil())
				defer r.Body.Close()
				Expect(r.StatusCode).To(Equal(http.StatusOK))
			})
		})

		Context("logged in", func() {
			BeforeEach(func() {
				getCookie = func() *http.Cookie { return cookies[0] }
				getURL = func() string { return fmt.Sprintf(addr, "gadgets", "", "") }
			})

			It("lets you get gadgets", func() {
				r := getReq()
				defer r.Body.Close()
				Expect(r.StatusCode).To(Equal(http.StatusOK))
				var g []quimby.Gadget
				dec := json.NewDecoder(r.Body)
				Expect(dec.Decode(&g)).To(BeNil())
				Expect(len(g)).To(Equal(2))
			})

			It("lets you get a gadget", func() {
				getURL = func() string { return fmt.Sprintf(addr, "gadgets/", sprinklers.Id, "") }
				r := getReq()
				defer r.Body.Close()

				var g quimby.Gadget
				dec := json.NewDecoder(r.Body)
				Expect(dec.Decode(&g)).To(BeNil())
				Expect(g.Name).To(Equal("sprinklers"))
				Expect(g.Host).To(Equal(ts.URL))
			})

			Context("notes", func() {
				BeforeEach(func() {
					getURL = func() string { return fmt.Sprintf(addr, "gadgets/", sprinklers.Id, "/notes") }
				})

				It("lets you add notes to a gadget", func() {

					getBuf = func() io.Reader {
						var buf bytes.Buffer
						enc := json.NewEncoder(&buf)
						note := map[string]string{
							"text": "jibber jabber",
						}
						enc.Encode(note)
						return &buf
					}
					getMethod = func() string { return "POST" }
					r := getReq()
					Expect(r.StatusCode).To(Equal(http.StatusOK))
					r.Body.Close()
				})

				It("lets you get notes from a gadget", func() {
					Expect(sprinklers.AddNote(quimby.Note{Text: "how's things?", Author: "me"}))

					r := getReq()
					Expect(r.StatusCode).To(Equal(http.StatusOK))
					var notes []quimby.Note
					dec := json.NewDecoder(r.Body)
					Expect(dec.Decode(&notes)).To(BeNil())
					r.Body.Close()
					Expect(len(notes)).To(Equal(1))
					n := notes[0]
					Expect(n.Text).To(Equal("how's things?"))
				})
			})

			It("gives a 404 for a non-gadget", func() {
				getURL = func() string { return fmt.Sprintf(addr, "gadgets/notarealid", "", "") }
				r := getReq()
				Expect(r.StatusCode).To(Equal(http.StatusNotFound))
				r.Body.Close()
			})

			Context("adding and deleting", func() {
				var (
					numGadgets int
				)

				BeforeEach(func() {
					getMethod = func() string { return "DELETE" }
					getURL = func() string { return fmt.Sprintf(addr, "gadgets/", sprinklers.Id, "") }
				})

				AfterEach(func() {
					getURL = func() string { return fmt.Sprintf(addr, "gadgets", "", "") }
					getMethod = func() string { return "GET" }
					getBuf = func() io.Reader { return nil }
					r := getReq()
					defer r.Body.Close()
					Expect(r.StatusCode).To(Equal(http.StatusOK))
					var g []quimby.Gadget
					dec := json.NewDecoder(r.Body)

					Expect(dec.Decode(&g)).To(BeNil())
					Expect(len(g)).To(Equal(numGadgets))
				})

				It("lets you delete a gadget", func() {
					r := getReq()
					r.Body.Close()
					numGadgets = 1
				})

				It("doesn't let a read only user delete a gadget", func() {
					getCookie = func() *http.Cookie { return readCookies[0] }
					r := getReq()
					Expect(r.StatusCode).To(Equal(http.StatusUnauthorized))
					r.Body.Close()
					numGadgets = 2
				})

				It("lets you add a gadget", func() {
					getBuf = func() io.Reader {
						var buf bytes.Buffer
						enc := json.NewEncoder(&buf)
						g := quimby.Gadget{
							Name: "back yard",
							Host: "http://overthere.com",
						}
						enc.Encode(g)
						return &buf
					}
					getURL = func() string { return fmt.Sprintf(addr, "gadgets/", "", "") }
					getMethod = func() string { return "POST" }
					r := getReq()
					r.Body.Close()
					Expect(r.StatusCode).To(Equal(http.StatusOK))
					numGadgets = 3
				})
			})

			It("lets you get the status of a gadget", func() {

				getURL = func() string { return fmt.Sprintf(addr, "gadgets/", sprinklers.Id, "/status") }
				getCookie = func() *http.Cookie { return readCookies[0] }
				r := getReq()
				defer r.Body.Close()

				m := map[string]map[string]gogadgets.Value{}
				dec := json.NewDecoder(r.Body)
				Expect(dec.Decode(&m)).To(BeNil())
				Expect(len(m)).To(Equal(2))
				v, ok := m["back yard"]["sprinklers"]
				Expect(ok).To(BeTrue())
				Expect(v.Value).To(BeFalse())
			})

			It("lets you send a command to a gadget", func() {
				getBuf = func() io.Reader {
					var buf bytes.Buffer
					enc := json.NewEncoder(&buf)
					m := map[string]string{
						"command": "turn on back yard sprinklers",
					}
					enc.Encode(m)
					return &buf
				}
				getMethod = func() string { return "POST" }
				getURL = func() string { return fmt.Sprintf(addr, "gadgets/", sprinklers.Id, "/command") }

				r := getReq()
				defer r.Body.Close()

				Eventually(func() int {
					return len(msgs)
				}).Should(Equal(1))

				msg := msgs[0]
				Expect(msg.Body).To(Equal("turn on back yard sprinklers"))
			})

			It("lets you get the value of a device", func() {
				getURL = func() string {
					return fmt.Sprintf(
						addr,
						"gadgets/",
						sprinklers.Id,
						"/locations/back%20yard/devices/sprinklers/status",
					)
				}
				r := getReq()
				defer r.Body.Close()

				var v gogadgets.Value
				dec := json.NewDecoder(r.Body)
				Expect(dec.Decode(&v)).To(BeNil())
				Expect(v.Value).To(BeTrue())
			})

			// XIt("lets you get a brewtoad method", func() {

			// 	u := fmt.Sprintf(
			// 		beerAddr,
			// 		"3-floyds-zombie-dust-clone?grain_temperature=70.0",
			// 	)
			// 	req, err := http.NewRequest("GET", u, nil)
			// 	Expect(err).To(BeNil())
			// 	req.AddCookie(cookies[0])
			// 	r, err := http.DefaultClient.Do(req)
			// 	Expect(err).To(BeNil())
			// 	defer r.Body.Close()

			// 	dec := json.NewDecoder(r.Body)
			// 	var method []string
			// 	Expect(dec.Decode(&method)).To(BeNil())
			// 	Expect(len(method)).To(Equal(37))
			// 	Expect(method[0]).To(Equal("fill hlt to 7.0 gallons"))
			// 	Expect(method[36]).To(Equal("stop filling carboy"))
			// })

			It("saves and gets stats", func() {

				var getVal floatGetter

				getBuf = func() io.Reader {
					v := map[string]float64{"value": getVal()}
					var buf bytes.Buffer
					enc := json.NewEncoder(&buf)
					enc.Encode(v)
					return &buf
				}

				getVal = func() float64 { return 33.3 }
				getURL = func() string {
					return fmt.Sprintf(
						addr2,
						"internal/gadgets/",
						sprinklers.Id,
						"/sources/front%20yard%20temperature",
					)
				}
				getMethod = func() string { return "POST" }
				n := time.Now()

				r := getReq()
				defer r.Body.Close()
				Expect(r.StatusCode).To(Equal(http.StatusOK))
				time.Sleep(10 * time.Millisecond)

				getVal = func() float64 { return 33.5 }
				r = getReq()
				defer r.Body.Close()

				getURL = func() string {
					return fmt.Sprintf(
						addr,
						"gadgets/",
						sprinklers.Id,
						fmt.Sprintf(
							"/sources/front%%20yard%%20temperature?start=%s&end=%s",
							url.QueryEscape(n.Add(-60*time.Second).Format(time.RFC3339Nano)),
							url.QueryEscape(n.Add(time.Second).Format(time.RFC3339Nano)),
						),
					)
				}

				getMethod = func() string { return "GET" }
				r = getReq()
				defer r.Body.Close()

				var points []quimby.DataPoint
				dec := json.NewDecoder(r.Body)
				Expect(dec.Decode(&points)).To(BeNil())
				Expect(len(points)).To(Equal(2))

				p1 := points[0]
				Expect(p1.Value).To(Equal(33.3))

				p2 := points[1]
				Expect(p2.Value).To(Equal(33.5))

				//Get them as csv
				getAccept = func() string { return "application/csv" }
				r = getReq()
				defer r.Body.Close()
				c := csv.NewReader(r.Body)
				rows, err := c.ReadAll()
				Expect(err).To(BeNil())
				Expect(len(rows)).To(Equal(3))

				h := rows[0]
				Expect(h[0]).To(Equal("time"))
				Expect(h[1]).To(Equal("value"))

				r1 := rows[1]
				Expect(r1[1]).To(Equal("33.300000"))

				r2 := rows[2]
				Expect(r2[1]).To(Equal("33.500000"))

			})

			Describe("websockets", func() {
				var (
					ws *websocket.Conn
				)

				BeforeEach(func() {
					a := strings.Replace(addr, "http", "ws", -1)
					u := fmt.Sprintf(
						a,
						"gadgets/",
						sprinklers.Id,
						"/websocket",
					)

					h := http.Header{"Origin": {u}, "Authorization": {token}}
					var err error
					ws, _, err = dialer.Dial(u, h)
					Expect(err).To(BeNil())
				})

				AfterEach(func() {
					ws.Close()
				})

				It("allows the sending of a message with a websocket", func() {

					uuid := gogadgets.GetUUID()
					msg := gogadgets.Message{
						Type:   gogadgets.COMMAND,
						Body:   "turn on back yard sprinklers",
						UUID:   uuid,
						Sender: "cli",
					}
					d, _ := json.Marshal(msg)
					err := ws.WriteMessage(websocket.TextMessage, d)
					Expect(err).To(BeNil())

					Eventually(func() int {
						return len(msgs)
					}).Should(Equal(1))

					msg = msgs[0]
					Expect(msg.Body).To(Equal("turn on back yard sprinklers"))
					Expect(msg.UUID).To(Equal(uuid))
				})

				It("allows the getting of a message with a websocket", func() {

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
					req, err := http.NewRequest("POST", fmt.Sprintf(addr2, "internal/updates", "", ""), &buf)
					Expect(err).To(BeNil())
					req.AddCookie(cookies[0])
					r, err := http.DefaultClient.Do(req)
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

			Context("on and off", func() {
				var (
					val bool
				)

				BeforeEach(func() {
					val = false
					getBuf = func() io.Reader {
						var buf bytes.Buffer
						enc := json.NewEncoder(&buf)
						v := gogadgets.Value{
							Value: val,
						}
						enc.Encode(v)
						return &buf
					}

					getURL = func() string {
						return fmt.Sprintf(
							addr,
							"gadgets/",
							sprinklers.Id,
							"/locations/front%20yard/devices/sprinklers/status",
						)
					}
					getMethod = func() string { return "POST" }
				})

				It("lets you turn on a device", func() {
					val = true
					r := getReq()
					defer r.Body.Close()

					Expect(len(msgs)).To(Equal(1))
					msg := msgs[0]
					Expect(msg.Body).To(Equal("turn on front yard sprinklers"))
				})

				It("lets you turn off a device", func() {
					val = false
					r := getReq()
					defer r.Body.Close()

					Expect(len(msgs)).To(Equal(1))
					msg := msgs[0]
					Expect(msg.Body).To(Equal("turn off front yard sprinklers"))
				})
			})
		})
	})

	Context("not logged in", func() {

		AfterEach(func() {
			getURL = func() string { return fmt.Sprintf(addr, "gadgets", "", "") }
			getMethod = func() string { return "GET" }
			getBuf = func() io.Reader { return nil }
			r := getReq()
			defer r.Body.Close()
			Expect(r.StatusCode).To(Equal(http.StatusOK))
			var g []quimby.Gadget
			dec := json.NewDecoder(r.Body)

			Expect(dec.Decode(&g)).To(BeNil())
			Expect(len(g)).To(Equal(2))
		})

		Describe("does not let you", func() {
			It("get gadgets", func() {
				u := fmt.Sprintf(addr, "gadgets", "", "")
				r, err := http.Get(u)
				Expect(err).To(BeNil())
				Expect(r.StatusCode).To(Equal(http.StatusUnauthorized))
			})
			It("get a gadget", func() {
				r, err := http.Get(fmt.Sprintf(addr, "gadgets/", sprinklers.Id, ""))
				Expect(err).To(BeNil())
				Expect(r.StatusCode).To(Equal(http.StatusUnauthorized))
				d, _ := ioutil.ReadAll(r.Body)
				Expect(strings.TrimSpace(string(d))).To(Equal("Not Authorized"))
				r.Body.Close()
			})

			It("delete a gadget", func() {
				req, err := http.NewRequest("DELETE", fmt.Sprintf(addr, "gadgets/", sprinklers.Id, ""), nil)
				Expect(err).To(BeNil())
				r, err := http.DefaultClient.Do(req)
				Expect(err).To(BeNil())
				Expect(r.StatusCode).To(Equal(http.StatusUnauthorized))
				d, _ := ioutil.ReadAll(r.Body)
				Expect(strings.TrimSpace(string(d))).To(Equal("Not Authorized"))
				defer r.Body.Close()
			})

			It("add a gadget", func() {
				var buf bytes.Buffer
				enc := json.NewEncoder(&buf)
				g := quimby.Gadget{
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

			It("get the status of a gadget", func() {
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

			It("send a command to a gadget", func() {
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

			It("does not turn on a device", func() {
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
