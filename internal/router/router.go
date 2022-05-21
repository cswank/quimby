package router

import (
	"log"
	"net/http"

	"github.com/cswank/quimby/internal/clients"
	"github.com/cswank/quimby/internal/config"
	"github.com/cswank/quimby/internal/errors"
	"github.com/cswank/quimby/internal/schema"
	"github.com/cswank/quimby/internal/templates"
	"github.com/go-chi/chi"
	"github.com/sec51/twofactor"
	"golang.org/x/crypto/bcrypt"
)

type (
	renderer interface {
		Name() string
		AddScripts([]string)
		AddLinks([]templates.Link)
		AddStylesheets([]string)
		Template() string
	}

	handler    func(http.ResponseWriter, *http.Request) error
	renderFunc func(http.ResponseWriter, *http.Request) (renderer, error)

	loginPage struct {
		templates.Page
		Error string
	}

	server struct {
		gadgets gadget
		user    user
		auth    auth
		hc      homekit
		clients *clients.Clients
		cfg     config.Config

		pages map[string]templates.Page
	}

	auth interface {
		Auth(h http.Handler) http.Handler
		GenerateCookie(username string) (*http.Cookie, error)
	}

	user interface {
		Get(username string) (schema.User, error)
	}

	homekit interface {
		Update(msg schema.Message)
	}

	gadget interface {
		GetAll() ([]schema.Gadget, error)
		Get(id int) (schema.Gadget, error)
	}
)

var (
	letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
)

func Serve(cfg config.Config, g gadget, u user, a auth, hc homekit) error {
	s := server{
		cfg:     cfg,
		gadgets: g,
		user:    u,
		auth:    a,
		hc:      hc,
		clients: clients.New(),
		pages: map[string]templates.Page{
			"login":  templates.NewPage("Quimby", "login.ghtml"),
			"logout": templates.NewPage("Quimby", "logout.ghtml"),
		},
	}

	pub := chi.NewRouter()
	priv := chi.NewRouter()

	pub.Route("/login", func(r chi.Router) {
		r.Get("/", handle(s.login))
		r.Post("/", handle(s.doLogin))
	})

	pub.Route("/logout", func(r chi.Router) {
		r.Get("/", handle(s.logout))
		r.Post("/", handle(s.doLogout))
	})

	pub.Get("/", s.redirect)
	pub.Get("/static/*", handle(s.static()))

	pub.With(s.auth.Auth).Route("/gadgets", func(r chi.Router) {
		r.Get("/", handle(s.getAll))
		r.Get("/{id}", handle(s.get))
		r.Get("/{id}/websocket", handle(s.connect))
		r.Get("/{id}/method", handle(s.method))
		r.Post("/{id}/method", handle(s.runMethod))
	})

	priv.Post("/status", handle(s.update))

	go func(r chi.Router) {
		if err := http.ListenAndServe(":3334", r); err != nil {
			log.Fatalf("unable to start private server: %s", err)
		}
	}(priv)

	return http.ListenAndServe(":3333", pub)
}

func (s *server) login(w http.ResponseWriter, req *http.Request) error {
	//	Error: req.URL.Query().Get("error"),
	return render(s.pages["login"], w, req)
}

func (s *server) logout(w http.ResponseWriter, req *http.Request) error {
	return render(s.pages["logout"], w, req)
}

func (s *server) doLogin(w http.ResponseWriter, req *http.Request) error {
	if err := req.ParseForm(); err != nil {
		return err
	}

	username := req.Form.Get("username")
	pw := req.Form.Get("password")
	token := req.Form.Get("token")

	usr, err := s.user.Get(username)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword(usr.Password, []byte(pw)); err != nil {
		return err
	}

	otp, err := twofactor.TOTPFromBytes(usr.TFA, "quimby")
	if err != nil {
		return err
	}

	if err := otp.Validate(token); err != nil {
		return errors.NewErrUnauthorized(err)
	}

	cookie, err := s.auth.GenerateCookie(username)
	if err != nil {
		return err
	}

	http.SetCookie(w, cookie)
	http.Redirect(w, req, "/gadgets", http.StatusSeeOther)
	return nil
}

func (s *server) doLogout(w http.ResponseWriter, req *http.Request) error {
	cookie := &http.Cookie{
		Name:   "quimby",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
	http.Redirect(w, req, "/login", http.StatusSeeOther)
	return nil
}

func handle(h handler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		err := h(w, req)
		if err == nil {
			return
		}

		log.Println(err)
		if errors.IsUnauthorized(err) {
			http.Redirect(w, req, `/login?error="invalid login"`, http.StatusSeeOther)
		}
	}
}

func render(pg templates.Page, w http.ResponseWriter, req *http.Request) error {
	t, scripts, stylesheets := templates.Get(pg.Template())
	pg.AddScripts(scripts)
	pg.AddStylesheets(stylesheets)
	pg.AddLinks([]templates.Link{{Name: "logout", Link: "/logout"}})
	return t.ExecuteTemplate(w, "base", pg)
}
