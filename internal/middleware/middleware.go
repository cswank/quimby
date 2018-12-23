package middleware

import (
	"log"
	"net/http"
)

type Handler func(http.ResponseWriter, *http.Request) error

func Handle(h Handler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		err := h(w, req)
		if err != nil {
			log.Println(err)
		}
	}
}
