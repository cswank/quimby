package quimby

import (
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/cswank/gogadgets"
)

var (
	pth = os.Getenv("QUIMBY_HOMEKIT_PATH")
)

type HomeKit struct {
	db          *bolt.DB
	id          string
	switches    map[string]accessory.Switch
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
	fmt.Println(user, pth)
	if user == "" || pth == "" {
		LG.Println("didn't set QUIMBY_USER or QUIMBY_HOMEKIT_PATH, homekit exiting")
		return
	}
	h.getSwitches()

	cfg := hc.Config{
		StoragePath: pth,
		Pin:         h.key,
	}

	var t hc.Transport
	var err error
	if len(h.accessories) == 1 {
		t, err = hc.NewIPTransport(cfg, h.accessories[0])
	} else if len(h.accessories) > 1 {
		t, err = hc.NewIPTransport(cfg, h.accessories[0], h.accessories[1:]...)
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
	h.switches = map[string]accessory.Switch{}
	h.cmds = []cmd{}
	h.accessories = []*accessory.Accessory{}
	for _, g := range gadgets {
		Register(g)
		if err := g.Fetch(); err != nil {
			log.Printf("not adding %s to homekit: %s\n", g.Name, err)
			continue
		}
		devices, err := g.Status()
		if err != nil {
			log.Printf("not adding %s to homekit: %s\n", g.Name, err)
			continue
		}
		for name, dev := range devices {
			if dev.Info.Direction == "output" {
				info := accessory.Info{
					Name:         name,
					Manufacturer: "gogadgets",
				}
				s := accessory.NewSwitch(info)
				h.switches[name] = *s
				h.cmds = append(h.cmds, newCMD(s, g, name))
				h.accessories = append(h.accessories, s.Accessory)
			}
		}
	}
}

type cmd struct {
	s   *accessory.Switch
	g   Gadget
	k   string
	on  string
	off string
	ch  chan gogadgets.Message
}

func newCMD(s *accessory.Switch, g Gadget, k string) cmd {
	c := cmd{
		s:   s,
		g:   g,
		k:   k,
		on:  fmt.Sprintf("turn on %s", k),
		off: fmt.Sprintf("turn off %s", k),
	}
	c.s.Switch.On.OnValueRemoteUpdate(func(on bool) {
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
			c.s.Switch.On.SetValue(msg.Value.Value.(bool))
		}
	}
}
