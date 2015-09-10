package controllers

import (
	"log"
	"os"

	"github.com/boltdb/bolt"
	"github.com/cswank/quimby/models"
)

var (
	db *bolt.DB
)

func init() {
	pth := os.Getenv("QUIMBY_DB")
	if pth == "" {
		log.Fatal("could not get a path to the db, please set the QUIMBY_DB var")
	}
	var err error
	db, err = models.GetDB(pth)
	if err != nil {
		log.Fatal("could open the db", err)
	}
}
