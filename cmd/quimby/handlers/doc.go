package handlers

import (
	"fmt"
	"net/http"
)

func GadgetOptions(w http.ResponseWriter, req *http.Request) {
	args := GetArgs(req)
	w.Write(
		[]byte(
			fmt.Sprintf(`{
  "POST": {
    "description": "send a command to the gogadgets target",
    "body": "{\"command\": \"turn on kitchen light\"}",
    "response": "no body"
  },
  "GET": {
    "description": "get the metadata for a gogadgets target",
    "response": "{\"host\": \"%s\", \"id\": \"%s\", \"name\": \"%s\"}"
  }
}`,
				args.Gadget.Host,
				args.Gadget.Id,
				args.Gadget.Name,
			),
		),
	)
}
