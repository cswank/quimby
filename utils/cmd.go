package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
	"github.com/cswank/quimby/models"
)

func SendCommand(db *bolt.DB) {
	gadgets, err := models.GetGadgets(db)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("pick a number")
	listGadgets(gadgets)
	var n int
	fmt.Scanf("%d", &n)
	g := gadgets[n-1]
	g.DB = db

	fmt.Print("command: ")

	in := bufio.NewReader(os.Stdin)

	cmd, err := in.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	log.Println(g.SendCommand(cmd))
}
