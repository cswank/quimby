package webapp

import (
	"fmt"
	"net/http"

	"github.com/cswank/quimby/cmd/quimby/handlers"
	"github.com/gorilla/context"
)

func ChartSetupPage(w http.ResponseWriter, req *http.Request) {
	args := handlers.GetArgs(req)

	inputs := map[string]string{}
	s, err := args.Gadget.Status()

	if err != nil {
		context.Set(req, "error", err)
		return
	}

	for _, msg := range s {
		if msg.Info.Direction == "input" {
			inputs[fmt.Sprintf("%s %s", msg.Location, msg.Name)] = fmt.Sprintf("/api/gadgets/%s/sources/%s%%20%s", args.Gadget.Id, msg.Location, msg.Name)
		}
	}

	p := chartSetupPage{
		gadgetPage: gadgetPage{
			userPage: userPage{
				User:  args.User.Username,
				Admin: handlers.Admin(args),
				Links: []link{
					{"quimby", "/"},
					{args.Gadget.Name, fmt.Sprintf("/gadgets/%s", args.Gadget.Id)},
					{"chart-setup", fmt.Sprintf("/gadgets/%s/chart-setup.html", args.Gadget.Id)},
				},
			},
			Gadget: args.Gadget,
		},
		Inputs: inputs,
		Spans:  []string{"hour", "day", "week", "month"},
		Action: fmt.Sprintf("/gadgets/%s/chart.html", args.Gadget.Id),
	}
	templates["chart-setup.html"].template.ExecuteTemplate(w, "base", p)
}

func ChartPage(w http.ResponseWriter, req *http.Request) {
	args := handlers.GetArgs(req)
	span := args.Args.Get("span")
	if span == "" {
		span = "day"
	}
	summarize := args.Args.Get("summarize")
	if summarize == "" {
		summarize = "0"
	}
	links := []link{
		{"quimby", "/"},
		{args.Gadget.Name, fmt.Sprintf("/gadgets/%s", args.Gadget.Id)},
		{"chart", fmt.Sprintf("/gadgets/%s/chart.html", args.Gadget.Id)},
	}
	if args.Args.Get("from-setup") == "true" {
		links = append(links[:2], link{"chart-setup", fmt.Sprintf("/gadgets/%s/chart-setup.html", args.Gadget.Id)}, links[2])
	}
	sources := args.Args["source"]
	p := chartPage{
		gadgetPage: gadgetPage{
			userPage: userPage{
				User:  args.User.Username,
				Admin: handlers.Admin(args),
				Links: links,
				CSS:   []string{"/css/nv.d3.css"},
			},
			Gadget: args.Gadget,
		},
		Span:      span,
		Sources:   sources,
		Summarize: summarize,
	}
	templates["chart.html"].template.ExecuteTemplate(w, "base", p)
}
