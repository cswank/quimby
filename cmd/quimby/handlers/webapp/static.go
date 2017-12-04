package webapp

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/GeertJohan/go.rice"
	"github.com/cswank/gogadgets"
	"github.com/cswank/quimby"
	"github.com/cswank/quimby/cmd/quimby/handlers"
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
	for _, pth := range []string{"head.html", "base.html", "navbar.html", "links.html", "index.html", "gadget.html", "chart.html", "chart-setup.html", "chart-setup.js", "chart-input.html", "chart-input.js", "furnace.html", "base.js", "gadget.js", "method.js", "edit-method.html", "edit-method.js", "method.html", "furnace.js", "chart.js", "device.html", "edit-gadget.html", "edit-gadget.js", "edit-user.html", "edit-user.js", "delete.html", "delete.js", "password.html", "new-user.html", "qr-code.html", "admin.html", "login.html", "logout.html"} {
		s, err := box.String(pth)
		if err != nil {
			log.Fatal(err)
		}
		data[pth] = s
	}

	templates = map[string]tmpl{
		"index.html":       {files: []string{"index.html"}},
		"links.html":       {files: []string{"links.html"}},
		"gadget.html":      {files: []string{"gadget.html", "base.js", "gadget.js", "method.js", "method.html", "device.html"}},
		"edit-method.html": {files: []string{"edit-method.html", "edit-method.js"}},
		"chart.html":       {files: []string{"chart.html", "chart.js"}},
		"chart-setup.html": {files: []string{"chart-setup.html", "chart-setup.js"}},
		"chart-input.html": {files: []string{"chart-input.html", "chart-input.js"}},
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

type dropdown struct {
	Link string
	Text string
}

type action struct {
	Name   string
	URI    template.URL
	Method string
}

type userPage struct {
	User      string
	Admin     bool
	Links     []link
	CSS       []string
	Dropdowns []dropdown
}

type chartPage struct {
	gadgetPage
	Span      string
	Sources   []string
	Summarize string
}

type chartInputPage struct {
	gadgetPage
	Name string
	Key  string
	Back string
}

type chartInput struct {
	Value string
	Setup string
	Key   string
}

type chartSetupPage struct {
	gadgetPage
	Inputs map[string]chartInput
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

type usage struct {
	Name  string
	Day   string
	Month string
}

type furnacePage struct {
	gadgetPage
	States      []string
	Furnace     gogadgets.Message
	Thermometer gogadgets.Message
	SetPoint    float64
	HeatOnTime  string
	CoolOnTime  string
	Usage       []usage
}

func IndexPage(w http.ResponseWriter, req *http.Request) error {
	args := handlers.GetArgs(req)
	g, err := quimby.GetGadgets()
	if err != nil {
		return err
	}

	i := indexPage{
		Gadgets: g,
		userPage: userPage{
			User:  args.User.Username,
			Admin: handlers.Admin(args),
			Links: []link{
				{"quimby", "/home"},
			},
		},
	}
	return templates["index.html"].template.ExecuteTemplate(w, "base", i)
}

func LinksPage(w http.ResponseWriter, req *http.Request) error {
	args := handlers.GetArgs(req)
	i := indexPage{
		userPage: userPage{
			User:  args.User.Username,
			Admin: handlers.Admin(args),
			Links: []link{
				{"quimby", "/home"},
			},
		},
	}
	return templates["links.html"].template.ExecuteTemplate(w, "base", i)
}

func displayValues(msg *gogadgets.Message) {
	if v, ok := msg.Value.Value.(float64); ok {
		msg.Value.Value = fmt.Sprintf("%.1f", v)
	}
}

func EditMethodPage(w http.ResponseWriter, req *http.Request) error {
	args := handlers.GetArgs(req)
	p := gadgetPage{
		userPage: userPage{
			User:  args.User.Username,
			Admin: handlers.Admin(args),
			Links: []link{
				{"quimby", "/home"},
				{args.Gadget.Name, fmt.Sprintf("/gadgets/%s", args.Gadget.Id)},
				{"method", fmt.Sprintf("/gadgets/%s/method.html", args.Gadget.Id)},
			},
		},
		Gadget: args.Gadget,
		URI:    fmt.Sprintf("/gadgets/%s", args.Gadget.Id),
	}
	return templates["edit-method.html"].template.ExecuteTemplate(w, "base", p)
}

func GadgetPage(w http.ResponseWriter, req *http.Request) error {
	args := handlers.GetArgs(req)
	s, err := args.Gadget.Status()

	if err != nil {
		return err
	}

	l := map[string][]gogadgets.Message{}
	for _, msg := range s {
		if msg.Sender == "method runner" || msg.Type == "method update" || strings.HasPrefix(msg.Sender, "xbee") {
			continue
		}
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
				{"quimby", "/home"},
				{args.Gadget.Name, fmt.Sprintf("/gadgets/%s", args.Gadget.Id)},
			},
			Dropdowns: []dropdown{
				{Link: fmt.Sprintf("/gadgets/%s/method.html", args.Gadget.Id), Text: "Method"},
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
			Usage:       getUsage(args.Gadget),
		}
		return templates["furnace.html"].template.ExecuteTemplate(w, "base", p)
	}
	return templates["gadget.html"].template.ExecuteTemplate(w, "base", g)
}

func getUsage(g *quimby.Gadget) []usage {
	var u []usage
	for _, name := range []string{"home furnace heat", "home furnace cool"} {
		parts := strings.Split(name, " ")
		n := parts[2]
		u = append(
			u,
			usage{
				Name:  n,
				Day:   getOnTime(name, -24*time.Hour, g),
				Month: getOnTime(name, -30*24*time.Hour, g),
			},
		)
	}
	return u
}

func getOnTime(name string, period time.Duration, g *quimby.Gadget) string {
	e := time.Now()
	s := e.Add(period)
	data, err := g.GetDataPoints(name, s, e, 0, true)
	if err != nil {
		return "0hr"
	}

	var d time.Duration
	var prev quimby.DataPoint
	end := len(data) - 1
	for i, p := range data {
		if p.Value > 0.5 && prev.Value < 0.5 {
			prev = p
		} else if p.Value < 0.5 && prev.Value > 0.5 {
			d += p.Time.Sub(prev.Time)
			prev = p
		} else if i == end && p.Value > 0.5 && prev.Value > 0.5 {
			d += p.Time.Sub(prev.Time)
		}
	}
	return parseDuration(d)
}
