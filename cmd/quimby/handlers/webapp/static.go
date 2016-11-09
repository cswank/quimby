package webapp

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/GeertJohan/go.rice"
	"github.com/cswank/gogadgets"
	"github.com/cswank/quimby"
	"github.com/cswank/quimby/cmd/quimby/handlers"
	"github.com/gorilla/context"
)

var (
	ErrPasswordsDoNotMatch = errors.New("passwords do not match")
	templates              map[string]tmpl
)

type tmpl struct {
	template *template.Template
	files    []string
}

func Init(box *rice.Box) {
	data := map[string]string{}
	for _, pth := range []string{"head.html", "base.html", "navbar.html", "links.html", "index.html", "gadget.html", "chart.html", "chart-setup.html", "furnace.html", "base.js", "gadget.js", "furnace.js", "chart.js", "device.html", "edit-gadget.html", "edit-gadget.js", "edit-user.html", "edit-user.js", "delete.html", "delete.js", "password.html", "new-user.html", "qr-code.html", "admin.html", "login.html", "logout.html"} {
		s, err := box.String(pth)
		if err != nil {
			log.Fatal(err)
		}
		data[pth] = s
	}

	templates = map[string]tmpl{
		"index.html":       {files: []string{"index.html"}},
		"links.html":       {files: []string{"links.html"}},
		"gadget.html":      {files: []string{"gadget.html", "base.js", "gadget.js", "device.html"}},
		"chart.html":       {files: []string{"chart.html", "chart.js"}},
		"chart-setup.html": {files: []string{"chart-setup.html"}},
		"furnace.html":     {files: []string{"furnace.html", "base.js", "furnace.js", "device.html"}},
		"edit-gadget.html": {files: []string{"edit-gadget.html", "edit-gadget.js"}},
		"edit-user.html":   {files: []string{"edit-user.html", "edit-user.js"}},
		"delete.html":      {files: []string{"delete.html", "delete.js"}},
		"password.html":    {files: []string{"password.html", "edit-user.js"}},
		"new-user.html":    {files: []string{"new-user.html", "edit-user.js"}},
		"qr-code.html":     {files: []string{"qr-code.html"}},
		"admin.html":       {files: []string{"admin.html"}},
		"login.html":       {files: []string{"login.html"}},
		"logout.html":      {files: []string{"logout.html", "edit-user.js"}},
	}

	base := []string{"head.html", "base.html", "navbar.html"}

	for key, val := range templates {
		t := template.New(key)
		var err error
		for _, f := range append(val.files, base...) {
			t, err = t.Parse(data[f])
			if err != nil {
				log.Fatal(err)
			}
		}
		val.template = t
		templates[key] = val
	}
}

type link struct {
	Name string
	Path string
}

type action struct {
	Name   string
	URI    template.URL
	Method string
}

type userPage struct {
	User  string
	Admin bool
	Links []link
	CSS   []string
}

type chartPage struct {
	gadgetPage
	Span      string
	Sources   []string
	Summarize string
}

type chartSetupPage struct {
	gadgetPage
	Inputs map[string]string
	Spans  []string
	Action string
}

type editUserPage struct {
	userPage
	EditUser    *quimby.User
	Permissions []string
	Actions     []action
	End         int
}

type confirmPage struct {
	userPage
	Resource string
	Name     string
	Actions  []action
	End      int
}

type indexPage struct {
	userPage
	Gadgets []quimby.Gadget
}

type adminPage struct {
	userPage
	Gadgets []link
	Users   []link
}

type gadgetPage struct {
	userPage
	Gadget    *quimby.Gadget
	Websocket template.URL
	Locations map[string][]gogadgets.Message
	Error     string
	URI       string
	Actions   []action
	End       int
}

type furnacePage struct {
	gadgetPage
	States      []string
	Furnace     gogadgets.Message
	Thermometer gogadgets.Message
	SetPoint    float64
}

func IndexPage(w http.ResponseWriter, req *http.Request) {
	args := handlers.GetArgs(req)
	g, err := quimby.GetGadgets(args.DB)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	i := indexPage{
		Gadgets: g,
		userPage: userPage{
			User:  args.User.Username,
			Admin: handlers.Admin(args),
			Links: []link{
				{"quimby", "/"},
			},
		},
	}
	templates["index.html"].template.ExecuteTemplate(w, "base", i)
}

func LinksPage(w http.ResponseWriter, req *http.Request) {
	args := handlers.GetArgs(req)
	i := indexPage{
		userPage: userPage{
			User:  args.User.Username,
			Admin: handlers.Admin(args),
			Links: []link{
				{"quimby", "/"},
			},
		},
	}
	templates["links.html"].template.ExecuteTemplate(w, "base", i)
}

func displayValues(msg *gogadgets.Message) {
	if v, ok := msg.Value.Value.(float64); ok {
		msg.Value.Value = fmt.Sprintf("%.1f", v)
	}
}

func GadgetPage(w http.ResponseWriter, req *http.Request) {
	args := handlers.GetArgs(req)
	s, err := args.Gadget.Status()

	if err != nil {
		context.Set(req, "error", err)
		return
	}

	l := map[string][]gogadgets.Message{}
	for _, msg := range s {
		msgs, ok := l[msg.Location]
		if !ok {
			msgs = []gogadgets.Message{}
		}
		displayValues(&msg)
		msgs = append(msgs, msg)
		l[msg.Location] = msgs
	}

	p := os.Getenv("QUIMBY_PORT")
	d := os.Getenv("QUIMBY_DOMAIN")
	var u string
	if p == "443" {
		u = fmt.Sprintf("wss://%s:%s", d, p)
	} else {
		u = fmt.Sprintf("ws://%s:%s", d, p)
	}

	g := gadgetPage{
		userPage: userPage{
			User:  args.User.Username,
			Admin: handlers.Admin(args),
			Links: []link{
				{"quimby", "/"},
				{args.Gadget.Name, fmt.Sprintf("/gadgets/%s", args.Gadget.Id)},
			},
		},
		Gadget:    args.Gadget,
		Websocket: template.URL(fmt.Sprintf("%s/api/gadgets/%s/websocket", u, args.Gadget.Id)),
		Locations: l,
		URI:       fmt.Sprintf("/gadgets/%s", args.Gadget.Id),
	}
	if args.Gadget.View == "furnace" {
		var f, t gogadgets.Message
		for _, m := range g.Locations["home"] {
			if m.Name == "furnace" {
				f = m
			} else {
				t = m
			}
		}
		var setPoint = 70.0
		if f.TargetValue != nil {
			setPoint = f.TargetValue.Value.(float64)
		}
		p := furnacePage{
			gadgetPage:  g,
			States:      []string{"heat", "cool", "off"},
			Furnace:     f,
			Thermometer: t,
			SetPoint:    setPoint,
		}
		templates["furnace.html"].template.ExecuteTemplate(w, "base", p)
	} else {
		templates["gadget.html"].template.ExecuteTemplate(w, "base", g)
	}
}
