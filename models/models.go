package models

import (
	"fmt"
	"os"
)

type Logger interface {
	Printf(string, ...interface{})
	Println(...interface{})
	Fatal(...interface{})
	Fatalf(string, ...interface{})
}

var (
	Clients *ClientHolder
	LG      Logger
	addr    string
	host    string
)

func GetAddr() string {
	host = os.Getenv("QUIMBY_HOST")
	if host == "" {
		LG.Println("please set QUIMBY_HOST")
	}
	if addr == "" {
		addr = fmt.Sprintf("%s:%s/internal/updates", os.Getenv("QUIMBY_HOST"), os.Getenv("QUIMBY_INTERNAL_PORT"))
	}
	return addr
}
