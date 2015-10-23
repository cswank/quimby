package controllers

import "fmt"

func GadgetOptions(args *Args) error {
	args.W.Write(
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
	return nil
}
