package models

import (
	"fmt"
	"os"
	"time"
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
	user    string
)

func GetAddr() string {
	host = os.Getenv("QUIMBY_HOST")
	if host == "" {
		LG.Println("please set QUIMBY_HOST")
	}
	user = os.Getenv("QUIMBY_USER")
	if user == "" {
		LG.Println("please set QUIMBY_USER")
	}
	if addr == "" {
		addr = fmt.Sprintf("%s:%s/internal/updates", os.Getenv("QUIMBY_HOST"), os.Getenv("QUIMBY_INTERNAL_PORT"))
	}
	return addr
}

func Register(g Gadget) error {
	token, err := GenerateToken(user, time.Duration(24*365*time.Hour))
	if err != nil {
		return err
	}
	_, err = g.Register(GetAddr(), token)
	return err
}
