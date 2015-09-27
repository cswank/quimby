package controllers

import (
	"github.com/boltdb/bolt"
	"github.com/cswank/gogadgets"
)

type Logger interface {
	Printf(string, ...interface{})
	Println(...interface{})
	Fatal(...interface{})
	Fatalf(string, ...interface{})
}

var (
	DB      *bolt.DB
	addr    string
	clients map[string]chan gogadgets.Message
	host    string
)
