package main_test

import (
	"bytes"

	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/cswank/gogadgets"
	"github.com/cswank/quimby"
	"github.com/gorilla/websocket"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	addr           = fmt.Sprintf("http://localhost:%s/api/%%s%%s%%s", os.Getenv("QUIMBY_PORT"))
	addr2          = fmt.Sprintf("http://localhost:%s/%%s%%s%%s", os.Getenv("QUIMBY_INTERNAL_PORT"))
	beerAddr       = fmt.Sprintf("http://localhost:%s/api/beer/%%s", os.Getenv("QUIMBY_PORT"))
	adminAddr      = fmt.Sprintf("http://localhost:%s/api/admin/%%s", os.Getenv("QUIMBY_PORT"))
	sprinklersId   = os.Getenv("QUIMBY_TEST_SPRINKLERS_ID")
	sprinklersHost = os.Getenv("QUIMBY_TEST_SPRINKLERS_HOST")
	token          string
	readToken      string
	adminToken     string
	cookie         *http.Cookie
	readCookie     *http.Cookie
	adminCookie    *http.Cookie
)

func addGadget(name string, host string) (string, error) {
	u := fmt.Sprintf(addr, "gadgets", "", "")
	buf := bytes.NewBufferString(fmt.Sprintf(`{"name": "%s", "host": "%s"}`, name, host))

	req, err := http.NewRequest("POST", u, buf)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", token)
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	p := r.Header.Get("Location")
	parts := strings.Split(p, "/")
	return parts[len(parts)-1], nil
}

func deleteGadgets() {
	u := fmt.Sprintf(addr, "gadgets", "", "")
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		log.Fatal("can't delete", err)
	}
	req.Header.Add("Authorization", token)
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("can't delete ", err)
	}

	defer r.Body.Close()
	var gs []quimby.Gadget
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&gs); err != nil {
		log.Fatal("can't delete ", err)
	}
	for _, g := range gs {
		if g.Id != sprinklersId {
			deleteItem(g.Id)
		}
	}
}

func deleteItem(id string) {
	u := fmt.Sprintf(addr, "gadgets/", id, "")
	req, err := http.NewRequest("DELETE", u, nil)
	if err != nil {
		log.Fatal("can't delete", err)
	}
	req.Header.Add("Authorization", token)
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("can't delete", err)
	}
	r.Body.Close()
}

func loginJWT(u string, p string) string {
	url := fmt.Sprintf(addr, "login?auth=jwt", "", "")
	buf := bytes.Buffer{}
	enc := json.NewEncoder(&buf)
	usr := &quimby.User{
		Username: u,
		Password: p,
	}
	enc.Encode(usr)

	r, err := http.Post(url, "application/json", &buf)
	if err != nil || r.StatusCode != http.StatusOK {
		log.Fatal("couldn't log in ", err)
	}
	return r.Header.Get("Authorization")
}

func loginCookie(u string, p string) *http.Cookie {
	url := fmt.Sprintf(addr, "login", "", "")
	buf := bytes.Buffer{}
	enc := json.NewEncoder(&buf)
	usr := &quimby.User{
		Username: u,
		Password: p,
	}
	enc.Encode(usr)
	r, err := http.Post(url, "application/json", &buf)
	if err != nil || r.StatusCode != http.StatusOK {
		log.Fatal("couldn't log in", err)
	}
	return r.Cookies()[0]
}

func initSprinklers() {
	u := fmt.Sprintf(addr, "gadgets/", sprinklersId, "/command")

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	m := map[string]string{
		"command": "turn on back yard sprinklers",
	}
	enc.Encode(m)

	req, err := http.NewRequest("POST", u, &buf)
	if err != nil {
		log.Fatal("error init sprinklers", err)
	}
	req.Header.Add("Authorization", token)
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("error init sprinklers", err)
	}
	r.Body.Close()

	buf = bytes.Buffer{}
	enc = json.NewEncoder(&buf)
	m = map[string]string{
		"command": "turn off back yard sprinklers",
	}
	enc.Encode(m)

	req, err = http.NewRequest("POST", u, &buf)
	if err != nil {
		log.Fatal("error init sprinklers", err)
	}
	req.Header.Add("Authorization", token)
	r, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("error init sprinklers", err)
	}
	r.Body.Close()
}

func init() {
	u := fmt.Sprintf(addr, "ping", "", "")
	for {
		_, err := http.Get(u)
		if err == nil {
			break
		}
	}
	token = loginJWT("me", "hushhush")
	readToken = loginJWT("him", "shhhhhhhh")
	adminToken = loginJWT("boss", "sosecret")

	cookie = loginCookie("me", "hushhush")
	readCookie = loginCookie("him", "shhhhhhhh")
	adminCookie = loginCookie("boss", "sosecret")
	initSprinklers()
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

var dialer = websocket.Dialer{
	Subprotocols:    []string{"p1", "p2"},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type requester func() *http.Response
type stringGetter func() string
type bufGetter func() io.Reader
type floatGetter func() float64
type cookieGetter func() *http.Cookie

var _ = Describe("Quimby", func() {
	var (
		getAccept stringGetter
		getReq    requester
		getURL    stringGetter
		getBuf    bufGetter
		getTok    stringGetter
		getMethod stringGetter
	)

	BeforeEach(func() {

		getMethod = func() string {
			return "GET"
		}

		getAccept = func() string { return "application/json" }

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
		deleteGadgets()
	})

	Describe("with jwt", func() {
		BeforeEach(func() {
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
					getTok = func() string { return readToken }
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

		Context("logged in", func() {
			It("lets you get the logged in user", func() {
				getURL = func() string { return fmt.Sprintf(addr, "currentuser", "", "") }
				r := getReq()
				defer r.Body.Close()
				Expect(r.StatusCode).To(Equal(http.StatusOK))
			})

			It("lets you get gadgets", func() {
				getURL = func() string { return fmt.Sprintf(addr, "gadgets", "", "") }
				r := getReq()
				defer r.Body.Close()
				Expect(r.StatusCode).To(Equal(http.StatusOK))
				var g []quimby.Gadget
				dec := json.NewDecoder(r.Body)
				Expect(dec.Decode(&g)).To(BeNil())
				Expect(len(g)).To(Equal(1))
			})

			It("lets you get a gadget", func() {
				getURL = func() string { return fmt.Sprintf(addr, "gadgets/", sprinklersId, "") }
				r := getReq()
				defer r.Body.Close()

				var g quimby.Gadget
				dec := json.NewDecoder(r.Body)
				Expect(dec.Decode(&g)).To(BeNil())
				Expect(g.Name).To(Equal("sprinklers"))
				Expect(g.Host).To(Equal(sprinklersHost))
			})

			It("lets you add notes to a gadget", func() {
				getURL = func() string { return fmt.Sprintf(addr, "gadgets/", sprinklersId, "/notes") }
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
					junkId, err := addGadget("junk", "http://localhost:4444")
					Expect(err).To(BeNil())
					getMethod = func() string { return "DELETE" }
					getURL = func() string { return fmt.Sprintf(addr, "gadgets/", junkId, "") }
					r := getReq()
					defer r.Body.Close()
					numGadgets = 1
				})

				It("doesn't let a read only user delete a gadget", func() {
					getTok = func() string { return readToken }
					getMethod = func() string { return "DELETE" }
					getURL = func() string { return fmt.Sprintf(addr, "gadgets/", sprinklersId, "") }
					r := getReq()
					r.Body.Close()
					Expect(r.StatusCode).To(Equal(http.StatusUnauthorized))
					numGadgets = 1
				})
			})

			Context("adding and updating", func() {
				var (
					newId string
				)

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
					u := r.Header.Get("Location")
					parts := strings.Split(u, "/")
					Expect(len(parts)).To(Equal(4))
					newId = parts[3]

					getMethod = func() string { return "GET" }
					getBuf = func() io.Reader { return nil }
					getURL = func() string { return fmt.Sprintf(addr, "gadgets/", newId, "") }

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
					newId = parts[3]

					getBuf = func() io.Reader {
						var buf bytes.Buffer
						enc := json.NewEncoder(&buf)
						g := quimby.Gadget{
							Id:   newId,
							Name: "back yard for reals",
							Host: "http://overthere.com",
						}
						enc.Encode(g)
						return &buf
					}

					getURL = func() string { return fmt.Sprintf(addr, "gadgets/", newId, "") }
					r = getReq()
					defer r.Body.Close()
					Expect(r.StatusCode).To(Equal(http.StatusOK))
					Expect(len(r.Header["Location"])).To(Equal(1))
					u = r.Header.Get("Location")
					Expect(u).To(Equal("/api/gadgets/" + newId))

					getURL = func() string { return fmt.Sprintf(addr, "gadgets/", newId, "") }
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
				getURL = func() string { return fmt.Sprintf(addr, "gadgets/", sprinklersId, "/status") }

				getTok = func() string { return readToken }
				r := getReq()
				defer r.Body.Close()

				Expect(r.StatusCode).To(Equal(http.StatusOK))

				var m map[string]gogadgets.Message
				dec := json.NewDecoder(r.Body)
				Expect(dec.Decode(&m)).To(BeNil())

				v := m["back yard sprinklers"]

				val, ok := v.Value.Value.(bool)
				Expect(ok).To(BeTrue())
				Expect(val).To(BeFalse())
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
				getURL = func() string { return fmt.Sprintf(addr, "gadgets/", sprinklersId, "/command") }
				r := getReq()
				defer r.Body.Close()
				Expect(r.StatusCode).To(Equal(http.StatusOK))

				Eventually(func() string {
					d, err := ioutil.ReadFile("/tmp/sprinklers.txt")
					Expect(err).To(BeNil())
					return string(d)
				}).Should(Equal("1"))

				getBuf = func() io.Reader {
					var buf bytes.Buffer
					enc := json.NewEncoder(&buf)
					m := map[string]string{
						"command": "turn off back yard sprinklers",
					}
					enc.Encode(m)
					return &buf
				}

				getMethod = func() string { return "POST" }
				getURL = func() string { return fmt.Sprintf(addr, "gadgets/", sprinklersId, "/command") }
				r = getReq()
				defer r.Body.Close()
				Expect(r.StatusCode).To(Equal(http.StatusOK))

				Eventually(func() string {
					d, err := ioutil.ReadFile("/tmp/sprinklers.txt")
					Expect(err).To(BeNil())
					return string(d)
				}).Should(Equal("0"))
			})

			It("lets you get the value of a device", func() {
				getURL = func() string {
					return fmt.Sprintf(
						addr,
						"gadgets/",
						sprinklersId,
						"/locations/back%20yard/devices/sprinklers/status",
					)
				}
				getMethod = func() string { return "GET" }

				r := getReq()
				defer r.Body.Close()
				Expect(r.StatusCode).To(Equal(http.StatusOK))

				var m gogadgets.Value
				dec := json.NewDecoder(r.Body)
				Expect(dec.Decode(&m)).To(BeNil())
				Expect(m.Value).To(BeFalse())
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
							sprinklersId,
							"/locations/back%20yard/devices/sprinklers/status",
						)
					}
					getMethod = func() string { return "POST" }
				})

				It("lets you turn a device on and off", func() {
					val = true
					r := getReq()
					r.Body.Close()
					Expect(r.StatusCode).To(Equal(http.StatusOK))

					Eventually(func() string {
						d, err := ioutil.ReadFile("/tmp/sprinklers.txt")
						Expect(err).To(BeNil())
						return string(d)
					}).Should(Equal("1"))

					val = false
					r = getReq()
					r.Body.Close()

					Eventually(func() string {
						d, err := ioutil.ReadFile("/tmp/sprinklers.txt")
						Expect(err).To(BeNil())
						return string(d)
					}).Should(Equal("0"))
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
						sprinklersId,
						"/sources/basement%20temperature",
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
						sprinklersId,
						fmt.Sprintf(
							"/sources/basement%%20temperature?start=%s&end=%s",
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
						sprinklersId,
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

					Eventually(func() string {
						d, err := ioutil.ReadFile("/tmp/sprinklers.txt")
						Expect(err).To(BeNil())
						return string(d)
					}).Should(Equal("1"))

					msg = gogadgets.Message{
						Type:   gogadgets.COMMAND,
						Body:   "turn off back yard sprinklers",
						UUID:   uuid,
						Sender: "cli",
					}
					d, _ = json.Marshal(msg)
					err = ws.WriteMessage(websocket.TextMessage, d)
					Expect(err).To(BeNil())

					Eventually(func() string {
						d, err := ioutil.ReadFile("/tmp/sprinklers.txt")
						Expect(err).To(BeNil())
						return string(d)
					}).Should(Equal("0"))

				})

				It("allows the getting of a message with a websocket", func() {

					uuid := gogadgets.GetUUID()
					msg := gogadgets.Message{
						Location: "back yard",
						Name:     "sprinklers",
						Type:     gogadgets.UPDATE,
						Host:     sprinklersHost,
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
			getCookie cookieGetter
		)

		BeforeEach(func() {

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
					return adminCookie
				}
				r := getReq()
				r.Body.Close()
				Expect(r.StatusCode).To(Equal(http.StatusOK))
			})

			It("does not let non-admins see how many websocket clients there are", func() {
				getCookie = func() *http.Cookie {
					return cookie
				}

				r := getReq()
				r.Body.Close()
				Expect(r.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("logging in and out", func() {
			BeforeEach(func() {
				getTok = func() string { return token }
			})
			It("lets you log in", func() {
				//Already logged in above in BeforeEach
				Expect(len(token)).ToNot(Equal(0))
			})

			It("lets you log out", func() {
				req, err := http.NewRequest("POST", fmt.Sprintf(addr, "logout", "", ""), nil)
				Expect(err).To(BeNil())
				req.AddCookie(cookie)
				r, err := http.DefaultClient.Do(req)
				Expect(err).To(BeNil())
				defer r.Body.Close()
				Expect(r.StatusCode).To(Equal(http.StatusOK))
			})
		})

		Context("logged in", func() {
			BeforeEach(func() {
				getCookie = func() *http.Cookie { return cookie }
				getURL = func() string { return fmt.Sprintf(addr, "gadgets", "", "") }
			})

			It("lets you get gadgets", func() {
				r := getReq()
				defer r.Body.Close()
				Expect(r.StatusCode).To(Equal(http.StatusOK))
				var g []quimby.Gadget
				dec := json.NewDecoder(r.Body)
				Expect(dec.Decode(&g)).To(BeNil())
				Expect(len(g)).To(Equal(1))
			})

			It("lets you get a gadget", func() {
				getURL = func() string { return fmt.Sprintf(addr, "gadgets/", sprinklersId, "") }
				r := getReq()
				defer r.Body.Close()

				var g quimby.Gadget
				dec := json.NewDecoder(r.Body)
				Expect(dec.Decode(&g)).To(BeNil())
				Expect(g.Name).To(Equal("sprinklers"))
				Expect(g.Host).To(Equal(sprinklersHost))
			})

			Context("notes", func() {
				var (
					startingNotes []quimby.Note
				)
				BeforeEach(func() {

					getURL = func() string { return fmt.Sprintf(addr, "gadgets/", sprinklersId, "/notes") }

					startingNotes = []quimby.Note{}
					getMethod = func() string { return "GET" }
					getBuf = func() io.Reader { return nil }
					r := getReq()
					Expect(r.StatusCode).To(Equal(http.StatusOK))
					dec := json.NewDecoder(r.Body)
					Expect(dec.Decode(&startingNotes)).To(BeNil())
				})

				It("lets you add  and get notes", func() {
					getBuf = func() io.Reader {
						var buf bytes.Buffer
						enc := json.NewEncoder(&buf)
						note := map[string]string{
							"text": "how's things?",
						}
						enc.Encode(note)
						return &buf
					}
					getMethod = func() string { return "POST" }
					r := getReq()
					Expect(r.StatusCode).To(Equal(http.StatusOK))
					r.Body.Close()

					getMethod = func() string { return "GET" }
					getBuf = func() io.Reader { return nil }
					r = getReq()
					Expect(r.StatusCode).To(Equal(http.StatusOK))
					var notes []quimby.Note
					dec := json.NewDecoder(r.Body)
					Expect(dec.Decode(&notes)).To(BeNil())
					r.Body.Close()
					Expect(len(notes)).To(Equal(len(startingNotes) + 1))
					n := notes[len(notes)-1]
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
					junkId, err := addGadget("junk", "http://localhost:4444")
					Expect(err).To(BeNil())
					getMethod = func() string { return "DELETE" }
					getURL = func() string { return fmt.Sprintf(addr, "gadgets/", junkId, "") }
					getTok = func() string { return token }
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
					getCookie = func() *http.Cookie { return readCookie }
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

				getURL = func() string { return fmt.Sprintf(addr, "gadgets/", sprinklersId, "/status") }
				getCookie = func() *http.Cookie { return readCookie }
				r := getReq()
				defer r.Body.Close()

				m := map[string]gogadgets.Message{}
				dec := json.NewDecoder(r.Body)
				Expect(dec.Decode(&m)).To(BeNil())
				Expect(len(m)).To(Equal(1))
				v, ok := m["back yard sprinklers"]
				Expect(ok).To(BeTrue())
				Expect(v.Value.Value).To(BeFalse())
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
				getURL = func() string { return fmt.Sprintf(addr, "gadgets/", sprinklersId, "/command") }

				r := getReq()
				r.Body.Close()
				Expect(r.StatusCode).To(Equal(http.StatusOK))

				Eventually(func() string {
					d, err := ioutil.ReadFile("/tmp/sprinklers.txt")
					Expect(err).To(BeNil())
					return string(d)
				}).Should(Equal("1"))

				getBuf = func() io.Reader {
					var buf bytes.Buffer
					enc := json.NewEncoder(&buf)
					m := map[string]string{
						"command": "turn off back yard sprinklers",
					}
					enc.Encode(m)
					return &buf
				}
				getMethod = func() string { return "POST" }
				getURL = func() string { return fmt.Sprintf(addr, "gadgets/", sprinklersId, "/command") }

				r = getReq()
				r.Body.Close()
				Expect(r.StatusCode).To(Equal(http.StatusOK))

				Eventually(func() string {
					d, err := ioutil.ReadFile("/tmp/sprinklers.txt")
					Expect(err).To(BeNil())
					return string(d)
				}).Should(Equal("0"))
			})

			It("lets you get the value of a device", func() {
				getURL = func() string {
					return fmt.Sprintf(
						addr,
						"gadgets/",
						sprinklersId,
						"/locations/back%20yard/devices/sprinklers/status",
					)
				}
				r := getReq()
				defer r.Body.Close()

				var v gogadgets.Value
				dec := json.NewDecoder(r.Body)
				Expect(dec.Decode(&v)).To(BeNil())
				Expect(v.Value).To(BeFalse())
			})

			// XIt("lets you get a brewtoad method", func() {

			// 	u := fmt.Sprintf(
			// 		beerAddr,
			// 		"3-floyds-zombie-dust-clone?grain_temperature=70.0",
			// 	)
			// 	req, err := http.NewRequest("GET", u, nil)
			// 	Expect(err).To(BeNil())
			// 	req.AddCookie(cookie)
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
						sprinklersId,
						"/sources/outside%20temperature",
					)
				}
				getMethod = func() string { return "POST" }
				n := time.Now()

				r := getReq()
				r.Body.Close()
				Expect(r.StatusCode).To(Equal(http.StatusOK))

				time.Sleep(100 * time.Millisecond)

				getVal = func() float64 { return 33.5 }
				r = getReq()
				r.Body.Close()
				Expect(r.StatusCode).To(Equal(http.StatusOK))

				getURL = func() string {
					return fmt.Sprintf(
						addr,
						"gadgets/",
						sprinklersId,
						fmt.Sprintf(
							"/sources/outside%%20temperature?start=%s&end=%s",
							url.QueryEscape(n.Add(-60*time.Second).Format(time.RFC3339Nano)),
							url.QueryEscape(n.Add(time.Second).Format(time.RFC3339Nano)),
						),
					)
				}

				getMethod = func() string { return "GET" }
				getBuf = func() io.Reader { return nil }
				var points []quimby.DataPoint

				Eventually(func() int {
					r = getReq()
					defer r.Body.Close()
					dec := json.NewDecoder(r.Body)
					Expect(dec.Decode(&points)).To(BeNil())
					return len(points)
				}).Should(Equal(2))

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
						sprinklersId,
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

					Eventually(func() string {
						d, err := ioutil.ReadFile("/tmp/sprinklers.txt")
						Expect(err).To(BeNil())
						return string(d)
					}).Should(Equal("1"))

					msg = gogadgets.Message{
						Type:   gogadgets.COMMAND,
						Body:   "turn off back yard sprinklers",
						UUID:   uuid,
						Sender: "cli",
					}
					d, _ = json.Marshal(msg)
					err = ws.WriteMessage(websocket.TextMessage, d)
					Expect(err).To(BeNil())

					Eventually(func() string {
						d, err := ioutil.ReadFile("/tmp/sprinklers.txt")
						Expect(err).To(BeNil())
						return string(d)
					}).Should(Equal("0"))
				})

				It("allows the getting of a message with a websocket", func() {

					uuid := gogadgets.GetUUID()
					msg := gogadgets.Message{
						Location: "back yard",
						Name:     "sprinklers",
						Type:     gogadgets.UPDATE,
						Host:     sprinklersHost,
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
					req.AddCookie(cookie)
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
							sprinklersId,
							"/locations/back%20yard/devices/sprinklers/status",
						)
					}
					getMethod = func() string { return "POST" }
				})

				It("lets you turn on a device", func() {
					val = true
					r := getReq()
					defer r.Body.Close()

					Eventually(func() string {
						d, err := ioutil.ReadFile("/tmp/sprinklers.txt")
						Expect(err).To(BeNil())
						return string(d)
					}).Should(Equal("1"))
				})

				It("lets you turn off a device", func() {
					val = false
					r := getReq()
					defer r.Body.Close()

					d, err := ioutil.ReadFile("/tmp/sprinklers.txt")
					Expect(err).To(BeNil())
					Expect(string(d)).To(Equal("0"))
				})
			})
		})
	})

	Context("not logged in", func() {

		BeforeEach(func() {
			getReq = func() *http.Response {
				req, err := http.NewRequest(getMethod(), getURL(), getBuf())
				Expect(err).To(BeNil())
				r, err := http.DefaultClient.Do(req)
				Expect(err).To(BeNil())
				return r
			}
		})

		AfterEach(func() {

			req, err := http.NewRequest("GET", fmt.Sprintf(addr, "gadgets", "", ""), nil)
			Expect(err).To(BeNil())
			req.Header.Add("Authorization", token)
			r, err := http.DefaultClient.Do(req)
			Expect(err).To(BeNil())
			defer r.Body.Close()
			Expect(r.StatusCode).To(Equal(http.StatusOK))
			var g []quimby.Gadget
			dec := json.NewDecoder(r.Body)

			Expect(dec.Decode(&g)).To(BeNil())
			Expect(len(g)).To(Equal(1))
		})

		Describe("does not let you", func() {
			It("get gadgets", func() {
				u := fmt.Sprintf(addr, "gadgets", "", "")
				r, err := http.Get(u)
				Expect(err).To(BeNil())
				Expect(r.StatusCode).To(Equal(http.StatusUnauthorized))
			})

			It("get a gadget", func() {
				r, err := http.Get(fmt.Sprintf(addr, "gadgets/", sprinklersId, ""))
				Expect(err).To(BeNil())
				Expect(r.StatusCode).To(Equal(http.StatusUnauthorized))
				d, _ := ioutil.ReadAll(r.Body)
				Expect(strings.TrimSpace(string(d))).To(Equal("Not Authorized"))
				r.Body.Close()
			})

			It("delete a gadget", func() {
				req, err := http.NewRequest("DELETE", fmt.Sprintf(addr, "gadgets/", sprinklersId, ""), nil)
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
					sprinklersId,
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
				u := fmt.Sprintf(addr, "gadgets/", sprinklersId, "")
				r, err := http.Post(u, "application/json", &buf)
				Expect(r.StatusCode).To(Equal(http.StatusUnauthorized))
				Expect(err).To(BeNil())
				d, _ := ioutil.ReadAll(r.Body)
				Expect(strings.TrimSpace(string(d))).To(Equal("Not Authorized"))
				r.Body.Close()
			})

			It("you to get of a websocket", func() {
				u := strings.Replace(
					fmt.Sprintf(
						addr,
						"gadgets/",
						sprinklersId,
						"/websocket",
					), "http", "ws", -1)
				h := http.Header{"Origin": {u}}
				ws, r, err := dialer.Dial(u, h)
				Expect(r.StatusCode).To(Equal(http.StatusUnauthorized))
				Expect(err).ToNot(BeNil())
				Expect(ws).To(BeNil())
			})

			It("turn on a device", func() {
				var buf bytes.Buffer
				enc := json.NewEncoder(&buf)
				v := gogadgets.Value{
					Value: true,
				}
				enc.Encode(v)

				u := fmt.Sprintf(
					addr,
					"gadgets/",
					sprinklersId,
					"/locations/front%20yard/devices/sprinklers/status",
				)
				r, err := http.Post(u, "application/json", &buf)
				Expect(err).To(BeNil())
				d, _ := ioutil.ReadAll(r.Body)
				Expect(strings.TrimSpace(string(d))).To(Equal("Not Authorized"))
				r.Body.Close()

				d2, err := ioutil.ReadFile("/tmp/sprinklers.txt")
				Expect(err).To(BeNil())
				Expect(string(d2)).To(Equal("0"))
			})

		})
	})
})
