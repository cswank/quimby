package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

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
	cmdMode   bool
	cmdTokens []rune
	links     [][]string
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

	g.Editor = ui.EditorFunc(commands)
	current = "nodes-cursor"
	g.SetLayout(layout)
	g.Cursor = true

	if err := g.MainLoop(); err != nil {
		if err != ui.ErrQuit {
			log.Fatal(err)
		}
	}
}

func commands(v *ui.View, key ui.Key, ch rune, mod ui.Modifier) {
	if ch == 'q' {
		current = "nodes-cursor"
		cli.Disconnect()
		return
	}
	if key == ui.KeyBackspace2 && len(cmdTokens) > 0 {
		cmdTokens = cmdTokens[0 : len(cmdTokens)-1]
		return
	}
	if key == ui.KeyEnter && len(cmdTokens) > 0 {
		cmd, err := getCmd()
		if err == nil {
			cli.Out <- gogadgets.Message{Type: gogadgets.COMMAND, Body: cmd}
		}
	} else if isNumber(ch) {
		cmdTokens = append(cmdTokens, ch)
	}
}

func isNumber(ch rune) bool {
	return strings.Index("0123456789", string(ch)) > -1
}

func getCmd() (string, error) {
	var cmd string
	for _, t := range cmdTokens {
		cmd += string(t)
	}
	i, err := strconv.ParseInt(cmd, 10, 64)
	if err != nil {
		return "", errors.New("invalid input")
	}
	cmdTokens = []rune{}
	if int(i)-1 >= len(links) {
		return "", errors.New("out of range")
	}
	return links[i-1][0], err
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
		v.Editable = true
		printNode(g)
	}
	return g.SetCurrentView(current)
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
	{"nodes-cursor", 'p', ui.ModNone, prev},
	{"nodes-cursor", ui.KeyArrowUp, ui.ModNone, prev},
	{"nodes-cursor", ui.KeyEnter, ui.ModNone, cmd},
	{"node", ui.KeyEsc, ui.ModNone, esc},
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
	printNode(g)
	return err
}

func prev(g *ui.Gui, v *ui.View) error {
	cx, cy := v.Cursor()
	if cy-1 < 0 {
		return nil
	}
	err := v.SetCursor(cx, cy-1)
	printNodes()
	printNode(g)
	return err
}

func cmd(g *ui.Gui, v *ui.View) error {
	current = "node"
	cmdTokens = []rune{}
	var err error
	v, err = g.View("node")
	if err != nil {
		return err
	}
	cmdMode = true

	_, cur := v.Cursor()
	if err := cli.Connect(cur, update); err != nil {
		return err
	}
	return v.SetCursor(0, 0)
}

func esc(g *ui.Gui, v *ui.View) error {
	current = "nodes-cursor"
	return nil
}

func printNode(g *ui.Gui) {
	v, _ := g.View("node")
	cv, _ := g.View("nodes-cursor")
	v.Clear()
	_, cur := cv.Cursor()
	n := cli.Nodes[cur]
	f := colors["color2"]
	i := 1
	links = [][]string{}
	var keys []string
	for key := range n.Devices {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		f(v, fmt.Sprintf("  %s", key))
		loc := n.Devices[key]
		var names []string
		for name := range loc {
			names = append(names, name)
		}
		sort.Strings(names)

		for _, name := range names {
			val := loc[name]
			var l string
			l, i = getLink(val, i)
			f(v, fmt.Sprintf("    %s: %s %s", name, getVal(val.Value), l))
			if l != "" {
				if val.Value.Value.(bool) {
					links = append(links, val.Info.Off)
				} else {
					links = append(links, val.Info.On)
				}
			}
		}
	}
}

func getLink(msg gogadgets.Message, i int) (string, int) {
	if msg.Info.Direction == "input" {
		return "", i
	}
	link := fmt.Sprintf("[%d]", i)
	i++
	return link, i
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
	u, err := url.Parse(*addr)
	if err != nil {
		log.Fatal(err)
	}
	p := filepath.Join(usr.HomeDir, ".qcli", u.Host)
	if exists(p) {
		d, err := ioutil.ReadFile(p)
		if err != nil {
			log.Fatal(err)
		}
		cli, err = quimby.NewClient(*addr, quimby.Token(string(d)))
		if err != nil {
			log.Fatal(err)
		}
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
		pw := gopass.GetPasswd()
		cli, err = quimby.NewClient(*addr)
		if err != nil {
			log.Fatal(err)
		}
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

func update(msg gogadgets.Message) {
	g.Execute(func(g *ui.Gui) error {
		printNode(g)
		return nil
	})
}
