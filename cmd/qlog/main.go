package main

import (
	"io"
	"log"
	"os"

	"github.com/cswank/quimby"
	ui "github.com/jroimartin/gocui"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	current   string
	hostLabel string
	g         *ui.Gui
	addr      = kingpin.Arg("addr", "quimby address").String()
	userArg   = kingpin.Flag("username", "username").Short('u').String()
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
	if os.Getenv("QUIMBY_USERNAME") != "" {
		username = os.Getenv("QUIMBY_USERNAME")
	} else if *userArg != "" {
		username = *userArg
	} else {
		log.Fatal("you must either set the QUIMBY_USERNAME env var or pass it in with the --username arg")
	}
	setupColors()
}

func main() {
	kingpin.Parse()
	cli, err = quimby.NewClient(*addr)
	if err != nil {
		log.Fatal(err)
	}

	login()
}
