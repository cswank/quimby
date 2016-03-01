package utils

import (
	"fmt"
	"log"
	"math"

	"github.com/cswank/quimby"
)

func GetToken() {
	var n string
	fmt.Printf("username: ")
	fmt.Scanf("%s", &n)
	token, err := quimby.GenerateToken(n, math.MaxInt64)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(token)
}
