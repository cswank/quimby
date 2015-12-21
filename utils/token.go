package utils

import (
	"fmt"
	"log"
	"math"

	"github.com/boltdb/bolt"
	"github.com/cswank/quimby/models"
)

func GetToken(db *bolt.DB) {
	var n string
	fmt.Printf("username: ")
	fmt.Scanf("%s", &n)
	token, err := models.GenerateToken(n, math.MaxInt64)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(token)
}
