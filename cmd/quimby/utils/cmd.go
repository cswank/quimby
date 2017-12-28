package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/cswank/quimby"
)

func SendCommand() {
	gadgets, err := quimby.GetGadgets()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("pick a number")
	listGadgets(gadgets)
	var n int
	fmt.Scanf("%d", &n)
	g := gadgets[n-1]

	fmt.Print("command: ")

	in := bufio.NewReader(os.Stdin)

	cmd, err := in.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	log.Println(g.SendCommand(cmd))
}
