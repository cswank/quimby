package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"strconv"

	"github.com/cswank/gogadgets"
	"github.com/cswank/quimby"
	"github.com/howeyc/gopass"
	ui "github.com/jroimartin/gocui"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	current   string
	hostLabel string
	g         *ui.Gui
	addr      = kingpin.Arg("addr", "quimby address").String()
	colors    map[string]func(io.Writer, string)
	username  string
	pw        string
	cli       *quimby.Client
	color1    string
	colorMap  map[string]string
)

func init() {
	username = os.Getenv("QUIMBY_USERNAME")
	setupColors()
}

func main() {
	kingpin.Parse()

	login()

	g = ui.NewGui()
	if err := g.Init(); err != nil {
		log.Panicln(err)
	}
	current = "cursor"
	defer g.Close()

	if err := keybindings(g); err != nil {
		log.Fatal(err)
	}

	g.SetLayout(layout)
	g.Cursor = true

	if err := g.MainLoop(); err != nil {
		if err != ui.ErrQuit {
			log.Fatal(err)
		}
	}
}

// exists returns whether the given file or directory exists or not
func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return !os.IsNotExist(err)
}

func layout(g *ui.Gui) error {
	x, y := g.Size()
	size := len(cli.Nodes)

	if v, err := g.SetView("nodes-label", -1, -1, len("nodes"), 1); err != nil {
		if err != ui.ErrUnknownView {
			return err
		}
		v.Frame = false
		fmt.Fprintln(v, hostLabel)
	}

	if v, err := g.SetView("nodes-cursor", 4, 0, 6, size+1); err != nil {
		if err != ui.ErrUnknownView {
			return err
		}
		v.Highlight = true
		v.Frame = false
	}

	if v, err := g.SetView("nodes", 6, 0, 20, size+1); err != nil {
		if err != ui.ErrUnknownView {
			return err
		}
		v.Frame = false
		printNodes()
	}

	if v, err := g.SetView("node", 20, 0, x, y-1); err != nil {
		if err != ui.ErrUnknownView {
			return err
		}
		v.Frame = false
		printNode()
	}
	return g.SetCurrentView("nodes-cursor")
}

type key struct {
	name string
	key  interface{}
	mod  ui.Modifier
	f    ui.KeybindingHandler
}

var keys = []key{
	{"", ui.KeyCtrlC, ui.ModNone, quit},
	{"", ui.KeyCtrlD, ui.ModNone, quit},
	{"nodes-cursor", ui.KeyCtrlN, ui.ModNone, next},
	{"nodes-cursor", 'n', ui.ModNone, next},
	{"nodes-cursor", ui.KeyArrowDown, ui.ModNone, next},
	{"nodes-cursor", ui.KeyCtrlP, ui.ModNone, prev},
	{"login", ui.KeyCtrlN, ui.ModNone, next},
}

func keybindings(g *ui.Gui) error {
	for _, k := range keys {
		if err := g.SetKeybinding(k.name, k.key, k.mod, k.f); err != nil {
			return err
		}
	}
	return nil
}

func quit(g *ui.Gui, v *ui.View) error {
	return ui.ErrQuit
}

func next(g *ui.Gui, v *ui.View) error {
	cx, cy := v.Cursor()
	if cy+1 >= len(cli.Nodes) {
		return nil
	}
	err := v.SetCursor(cx, cy+1)
	printNodes()
	printNode()
	return err
}

func prev(g *ui.Gui, v *ui.View) error {
	cx, cy := v.Cursor()
	if cy-1 < 0 {
		return nil
	}
	err := v.SetCursor(cx, cy-1)
	printNodes()
	printNode()
	return err
}

func printNode() {
	v, _ := g.View("node")
	cv, _ := g.View("nodes-cursor")
	v.Clear()
	_, cur := cv.Cursor()
	n := cli.Nodes[cur]
	f := colors["color2"]
	for key, loc := range n.Devices {
		f(v, fmt.Sprintf("  %s", key))
		for name, val := range loc {
			f(v, fmt.Sprintf("    %s: %s", name, getVal(val)))
		}
	}
}

func getVal(val gogadgets.Value) string {
	var s string
	switch val.Value.(type) {
	case bool:
		if val.Value.(bool) {
			s = green("on")
		} else {
			s = red("off")
		}
	case float64:
		s = strconv.FormatFloat(val.Value.(float64), 'f', 1, 64)
	}
	return s
}

func printNodes() {
	hv, _ := g.View("nodes")
	cv, _ := g.View("nodes-cursor")
	hv.Clear()
	_, cur := cv.Cursor()
	var i int
	for _, n := range cli.Nodes {
		f := colors["color1"]
		if i == cur {
			f = colors["color3"]
		}
		f(hv, n.Name)
		i++
	}
}

func setupColors() {
	colorMap = map[string]string{
		"black":   "30",
		"red":     "31",
		"green":   "32",
		"yellow":  "33",
		"blue":    "34",
		"magenta": "35",
		"cyan":    "36",
		"white":   "37",
	}

	color1 = os.Getenv("QCLI_COLOR_1")
	if color1 == "" {
		color1 = "green"
	}
	color2 := os.Getenv("QCLI_COLOR_2")
	if color2 == "" {
		color2 = "white"
	}
	color3 := os.Getenv("QCLI_COLOR_3")
	if color3 == "" {
		color3 = "yellow"
	}

	colors = map[string]func(io.Writer, string){
		"color1": func(w io.Writer, s string) {
			fmt.Fprintf(w, fmt.Sprintf("\033[%sm%%s\033[%sm\n", colorMap[color1], colorMap[color1]), s)
		},

		"color2": func(w io.Writer, s string) {
			fmt.Fprintf(w, fmt.Sprintf("\033[%sm%%s\033[%sm\n", colorMap[color2], colorMap[color1]), s)
		},

		"color3": func(w io.Writer, s string) {
			fmt.Fprintf(w, fmt.Sprintf("\033[%sm%%s\033[%sm\n", colorMap[color3], colorMap[color1]), s)
		},
	}
	hostLabel = fmt.Sprintf("\033[%smnodes\033[%sm\n", colorMap[color2], colorMap[color1])
}

func red(s string) string {
	return fmt.Sprintf(fmt.Sprintf("\033[%sm%%s\033[%sm", colorMap["red"], colorMap[color1]), s)
}

func green(s string) string {
	return fmt.Sprintf(fmt.Sprintf("\033[%sm%%s\033[%sm", colorMap["green"], colorMap[color1]), s)
}

func login() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	d := fmt.Sprintf("%s/.qcli", usr.HomeDir)
	p := fmt.Sprintf("%s/.qcli/token", usr.HomeDir)
	if exists(p) {
		d, err := ioutil.ReadFile(p)
		if err != nil {
			log.Fatal(err)
		}
		cli = quimby.NewClient(*addr, quimby.Token(string(d)))
		if err := cli.GetNodes(); err != nil {
			os.Remove(p)
			login()
		}
	} else {
		if !exists(d) {
			if os.Mkdir(d, 0700); err != nil {
				log.Fatal(err)
			}
		}
		fmt.Printf("password: ")
		pw, err := gopass.GetPasswd()
		if err != nil {
			log.Fatal(err)
		}
		cli = quimby.NewClient(*addr)
		token, err := cli.Login(username, string(pw))
		if err != nil {
			log.Fatal(err)
		}
		if err := ioutil.WriteFile(p, []byte(token), 0700); err != nil {
			log.Fatal(err)
		}

		if err := cli.GetNodes(); err != nil {
			log.Fatal(err)
		}
	}
}
