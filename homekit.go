package quimby

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/boltdb/bolt"
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/cswank/gogadgets"
)

var (
	pth = os.Getenv("QUIMBY_HOMEKIT_PATH")
)

type HomeKit struct {
	db  *bolt.DB
	id  string
	key string
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
}

func (h *HomeKit) getSwitches() {
	gadgets, err := GetGadgets(h.db)
	if err != nil {
		log.Fatal(err)
	}
	for _, g := range gadgets {
		p := path.Join(pth, g.Name)
		cfg := hc.Config{
			StoragePath: p,
			Pin:         h.key,
		}

		accessories := []*accessory.Accessory{}
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
				accessories = append(accessories, s.Accessory)
				connect(s, g, name)
			} else if dev.Info.Direction == "input" && dev.Info.Type == "thermometer" {
				info := accessory.Info{
					Name:         name,
					Manufacturer: "gogadgets",
				}
				t := accessory.NewTemperatureSensor(info, dev.Value.Value.(float64), 0.0, 212.0, 0.1)
				accessories = append(accessories, t.Accessory)
				connect(t, g, name)
			}
		}

		var t hc.Transport
		if len(accessories) == 1 {
			t, err = hc.NewIPTransport(cfg, accessories[0])
		} else if len(accessories) > 1 {
			t, err = hc.NewIPTransport(cfg, accessories[0], accessories[1:]...)
		} else {
			return
		}
		if err != nil {
			log.Println("not starting accesory", err)
		} else {
			go t.Start()
		}
	}
}

func connect(a interface{}, g Gadget, k string) {
	switch a.(type) {
	case *accessory.Switch:
		connectSwitch(a.(*accessory.Switch), g, k)
	case *accessory.Thermometer:
		connectThermometer(a.(*accessory.Thermometer), g, k)
	}
}

func connectThermometer(t *accessory.Thermometer, g Gadget, k string) {
	ch := make(chan gogadgets.Message)
	uuid := gogadgets.GetUUID()
	Clients.Add(g.Host, uuid, ch)
	go func(ch chan gogadgets.Message, k string, t *accessory.Thermometer) {
		for {
			msg := <-ch
			key := fmt.Sprintf("%s %s", msg.Location, msg.Name)
			if key == k {
				temp := (msg.Value.Value.(float64) - 32.0) / 1.8
				t.TempSensor.CurrentTemperature.SetValue(temp)
			}
		}
	}(ch, k, t)
}

func connectSwitch(s *accessory.Switch, g Gadget, k string) {
	s.Switch.On.OnValueRemoteUpdate(func(on bool) {
		if on == true {
			g.SendCommand(fmt.Sprintf("turn on %s", k))
		} else {
			g.SendCommand(fmt.Sprintf("turn off %s", k))
		}
	})
	ch := make(chan gogadgets.Message)
	uuid := gogadgets.GetUUID()
	Clients.Add(g.Host, uuid, ch)
	go func(ch chan gogadgets.Message, k string, s *accessory.Switch) {
		for {
			msg := <-ch
			key := fmt.Sprintf("%s %s", msg.Location, msg.Name)
			if key == k {
				s.Switch.On.SetValue(msg.Value.Value.(bool))
			}
		}
	}(ch, k, s)
}
