package quimby

import (
	"fmt"
	"log"
	"math"
	"os"
	"path"

	"github.com/boltdb/bolt"
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
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
	h.getDevices()
}

func (h *HomeKit) getDevices() {
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
				accessories = getOutputDevice(name, dev, g, accessories)
			} else if dev.Info.Direction == "input" {
				accessories = getInputDevice(name, dev, g, accessories)
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

func getInputDevice(name string, dev gogadgets.Message, g Gadget, accessories []*accessory.Accessory) []*accessory.Accessory {
	if dev.Info.Type != "thermometer" {
		return accessories
	}

	info := accessory.Info{
		Name:         name,
		Manufacturer: "gogadgets",
	}
	t := accessory.NewTemperatureSensor(info, dev.Value.Value.(float64), 0.0, 212.0, 0.1)
	connect(t, g, name)
	return append(accessories, t.Accessory)
}

func getOutputDevice(name string, dev gogadgets.Message, g Gadget, accessories []*accessory.Accessory) []*accessory.Accessory {
	if dev.Info.Type == "thermostat" {
		return getThermostat(name, dev, g, accessories)
	}
	info := accessory.Info{
		Name:         name,
		Manufacturer: "gogadgets",
	}
	s := accessory.NewSwitch(info)
	connect(s, g, name)
	return append(accessories, s.Accessory)
}

func getThermostat(name string, dev gogadgets.Message, g Gadget, accessories []*accessory.Accessory) []*accessory.Accessory {
	info := accessory.Info{
		Name:         name,
		Manufacturer: "gogadgets",
	}
	s := accessory.NewThermostat(info, 70.0, 0.0, 90.0, 0.1)
	connect(s, g, name)
	return append(accessories, s.Accessory)
}

func connect(a interface{}, g Gadget, k string) {
	switch a.(type) {
	case *accessory.Switch:
		connectSwitch(a.(*accessory.Switch), g, k)
	case *accessory.Thermometer:
		connectThermometer(a.(*accessory.Thermometer), g, k)
	case *accessory.Thermostat:
		connectThermostat(a.(*accessory.Thermostat), g, k)
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

func connectThermostat(t *accessory.Thermostat, g Gadget, k string) {
	//message from homekit
	t.Thermostat.TargetHeatingCoolingState.OnValueRemoteUpdate(func(state int) {
		t.Thermostat.TargetHeatingCoolingState.SetValue(state)
		switch state {
		case characteristic.TargetHeatingCoolingStateOff:
			g.SendCommand("turn off furnace")
		case characteristic.TargetHeatingCoolingStateHeat:
			v := t.Thermostat.TargetTemperature.GetValue()
			v = math.Floor(1.8*v + 32.0 + .5)
			g.SendCommand(fmt.Sprintf("heat home to %d F", int(v)))
		case characteristic.TargetHeatingCoolingStateCool:
			v := t.Thermostat.TargetTemperature.GetValue()
			v = math.Floor(1.8*v + 32.0 + .5)
			g.SendCommand(fmt.Sprintf("cool home to %d F", int(v)))
		}
	})

	t.Thermostat.TargetTemperature.OnValueRemoteUpdate(func(c float64) {
		f := int(math.Floor(1.8*c + 32.0 + .5))
		s := t.Thermostat.TargetHeatingCoolingState.GetValue()
		switch s {
		case characteristic.TargetHeatingCoolingStateOff:
			g.SendCommand("turn off furnace")
		case characteristic.TargetHeatingCoolingStateHeat:
			g.SendCommand(fmt.Sprintf("heat home to %d F", f))
		case characteristic.TargetHeatingCoolingStateCool:
			fmt.Println("sending command", g.SendCommand(fmt.Sprintf("cool home to %d F", f)))
		}
	})

	// ch := make(chan gogadgets.Message)
	// uuid := gogadgets.GetUUID()
	// Clients.Add(g.Host, uuid, ch)
	// go func(ch chan gogadgets.Message, k string, t *accessory.Thermostat) {
	// 	for {
	// 		msg := <-ch
	// 		key := fmt.Sprintf("%s %s", msg.Location, msg.Name)
	// 		if key == k {
	// 			fmt.Println("update from furnace", msg)
	// 			// if msg.Value.Cmd == "turn off home furnace" {
	// 			// 	t.Thermostat.TargetHeatingCoolingState.SetValue(characteristic.TargetHeatingCoolingStateOff)
	// 			// } else {

	// 			// case  "":
	// 			// 	fmt.Println("somebody turned on the furnace")
	// 		} // else if key == thermometer {
	// 		//fmt.Println("tell homekit thermostat that the temperature has changed")
	// 		//}
	// 	}
	// }(ch, k, t)
}
