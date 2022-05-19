package gadget

import (
	"fmt"
	"log"

	"github.com/cswank/quimby/internal/repository"
)

func Create(r *repository.Gadget, name, url string) {
	g, err := r.Create(name, url)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("created gadget %s\n", g)
}

func Delete(r *repository.Gadget, id int) {
	if err := r.Delete(id); err != nil {
		log.Fatal(err)
	}
}

func Update(r *repository.Gadget, id int, name, url string) {
	if err := r.Update(id, name, url); err != nil {
		log.Fatal(err)
	}
}

func List(r *repository.Gadget) {
	gds, err := r.List()
	if err != nil {
		log.Fatal(err)
	}

	for _, g := range gds {
		fmt.Println(g.String())
	}
}
