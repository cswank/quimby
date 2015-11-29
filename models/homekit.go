package models

import (
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
	"github.com/brutella/hc/hap"
	"github.com/brutella/hc/model"
	"github.com/brutella/hc/model/accessory"
	"github.com/cswank/gogadgets"
)

var (
	user = os.Getenv("QUIMBY_USER")
	pth  = os.Getenv("QUIMBY_HOMEKIT_PATH")
)

type HomeKit struct {
	db          *bolt.DB
	id          string
	switches    map[string]model.Switch
	accessories []*accessory.Accessory
	key         string
	cmds        []cmd
}

func NewHomeKit(key string, db *bolt.DB) *HomeKit {
	return &HomeKit{
		id:  "homekit",
		key: key,
		db:  db,
	}
}

func (h *HomeKit) Start() {
	if user == "" || pth == "" {
		LG.Println("didn't set QUIMBY_USER or QUIMBY_HOMEKIT_PATH, homekit exiting")
		return
	}
	h.getSwitches()
	var t hap.Transport
	var err error
	if len(h.accessories) == 1 {
		t, err = hap.NewIPTransport(h.key, pth, h.accessories[0])
	} else if len(h.accessories) > 1 {
		t, err = hap.NewIPTransport(h.key, pth, h.accessories[0], h.accessories[1:]...)
	} else {
		return
	}
	if err != nil {
		log.Fatal(err)
	}
	t.Start()
	LG.Println("homekit is done")
}

func (h *HomeKit) getSwitches() {
	h.cmds = []cmd{}
	gadgets, err := GetGadgets(h.db)
	if err != nil {
		log.Fatal(err)
	}
	h.switches = map[string]model.Switch{}
	h.cmds = []cmd{}
	h.accessories = []*accessory.Accessory{}
	for _, g := range gadgets {
		Register(g)
		if err := g.Fetch(); err != nil {
			log.Println("not adding %s to homekit: %s", g.Name, err)
			continue
		}
		devices, err := g.Status()
		if err != nil {
			log.Println("not adding %s to homekit: %s", g.Name, err)
			continue
		}
		for name, dev := range devices {
			if dev.Info.Direction == "output" {
				info := model.Info{
					Name:         name,
					Manufacturer: "gogadgets",
				}
				s := accessory.NewSwitch(info)

				h.switches[name] = s
				h.cmds = append(h.cmds, newCMD(s, g, name))
				h.accessories = append(h.accessories, s.Accessory)
			}
		}
	}
}

type cmd struct {
	s   model.Switch
	g   Gadget
	k   string
	on  string
	off string
	ch  chan gogadgets.Message
}

func newCMD(s model.Switch, g Gadget, k string) cmd {
	c := cmd{
		s:   s,
		g:   g,
		k:   k,
		on:  fmt.Sprintf("turn on %s", k),
		off: fmt.Sprintf("turn off %s", k),
	}
	c.s.OnStateChanged(func(on bool) {
		if on == true {
			c.g.SendCommand(c.on)
		} else {
			c.g.SendCommand(c.off)
		}
	})
	c.ch = make(chan gogadgets.Message)
	uuid := gogadgets.GetUUID()
	Clients.Add(g.Host, uuid, c.ch)
	go c.listen()
	return c
}

func (c *cmd) listen() {
	for {
		msg := <-c.ch
		key := fmt.Sprintf("%s %s", msg.Location, msg.Name)
		if key == c.k {
			c.s.SetOn(msg.Value.Value.(bool))
		}
	}
}
