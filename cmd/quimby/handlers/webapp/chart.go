package webapp

import (
	"fmt"
	"net/http"

	"github.com/cswank/quimby/cmd/quimby/handlers"
)

func ChartSetupPage(w http.ResponseWriter, req *http.Request) error {
	args := handlers.GetArgs(req)

	inputs := map[string]chartInput{}
	for _, name := range args.Gadget.GetDataPointSources() {
		inputs[name] = chartInput{
			Value: fmt.Sprintf("/api/gadgets/%s/sources/%s", args.Gadget.Id, name),
			Setup: fmt.Sprintf("/gadgets/%s/chart-setup/%s", args.Gadget.Id, name),
			Key:   fmt.Sprintf("%s %s", args.Gadget.Id, name),
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
	return templates["chart-setup.html"].template.ExecuteTemplate(w, "base", p)
}

func ChartInputPage(w http.ResponseWriter, req *http.Request) error {
	args := handlers.GetArgs(req)
	name := args.Vars["name"]
	p := chartInputPage{
		gadgetPage: gadgetPage{
			userPage: userPage{
				User:  args.User.Username,
				Admin: handlers.Admin(args),
				Links: []link{
					{"quimby", "/"},
					{args.Gadget.Name, fmt.Sprintf("/gadgets/%s", args.Gadget.Id)},
					{"chart-setup", fmt.Sprintf("/gadgets/%s/chart-setup.html", args.Gadget.Id)},
					{name, fmt.Sprintf("/gadgets/%s/chart-setup/%s", args.Gadget.Id, name)},
				},
			},
			Gadget: args.Gadget,
		},
		Name: args.Vars["name"],
		Key:  fmt.Sprintf("%s %s", args.Gadget.Id, args.Vars["name"]),
		Back: fmt.Sprintf("/gadgets/%s/chart-setup.html", args.Gadget.Id),
	}
	return templates["chart-input.html"].template.ExecuteTemplate(w, "base", p)
}

func ChartPage(w http.ResponseWriter, req *http.Request) error {
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
	return templates["chart.html"].template.ExecuteTemplate(w, "base", p)
}
