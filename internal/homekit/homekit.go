package homekit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/cswank/gogadgets"
	"github.com/kelseyhightower/envconfig"
)

type homekitState uint8

func (h homekitState) String() string {
	switch h {
	case 1:
		return "heat home"
	case 2:
		return "cool home"
	default:
		return "turn off home furnace"
	}
}

const (
	off  homekitState = 0
	heat homekitState = 1
	cool homekitState = 2
)

type (
	config struct {
		Pin         string `envconfig:"PIN" required:"true"`
		Port        string `envconfig:"PORT" required:"true"`
		FurnaceHost string `envconfig:"FURNACE_HOST" required:"true"`
	}

	Homekit struct {
		pin  string
		port string
		//the thermostat host
		furnaceHost string
		//the location + name id of the temperature sensor (must be in the same location)
		updateFurnace func(*gogadgets.Message) bool
		furnaceOn     func(*gogadgets.Value) error
		furnaceOff    func() error
	}
)

func New() (*Homekit, error) {
	var cfg config
	err := envconfig.Process("HOMEKIT", &cfg)
	if err != nil {
		return nil, err
	}

	h := &Homekit{
		pin:         cfg.Pin,
		port:        cfg.Port,
		furnaceHost: cfg.FurnaceHost,
		updateFurnace: func(*gogadgets.Message) bool {
			log.Println("not implemented")
			return false
		},
		furnaceOn: func(*gogadgets.Value) error {
			log.Println("not implemented")
			return nil
		},
		furnaceOff: func() error {
			log.Println("not implemented")
			return nil
		},
	}

	go h.start()

	return h, h.init()
}

func (h *Homekit) Update(msg gogadgets.Message) {
	log.Printf("%+v", msg)
}

func (h *Homekit) start() {
	bridge := accessory.NewBridge(accessory.Info{Name: "Quimby"})

	var ac []*accessory.Accessory
	if h.furnaceHost != "" {
		ac = append(ac, h.furnace())
	}

	tr, err := hc.NewIPTransport(
		hc.Config{Pin: h.pin, Port: h.port},
		bridge.Accessory,
		ac...,
	)

	if err != nil {
		log.Panic(err)
	}

	tr.Start()
}

func (h *Homekit) furnace() *accessory.Accessory {
	furnace := accessory.NewThermostat(accessory.Info{Name: "Thermostat"}, 20, 16, 26, 1)

	state := off
	furnace.Thermostat.TargetHeatingCoolingState.OnValueRemoteUpdate(func(i int) {
		state = homekitState(i)
		log.Printf("setting state: %s", state)
	})

	furnace.Thermostat.TargetTemperature.OnValueRemoteUpdate(func(c float64) {
		f := float64(c*1.8 + 32.0)
		log.Printf("setting temperature to %f, state: %s", f, state)
		msg := gogadgets.Message{Type: gogadgets.COMMAND, Sender: "homekit"}

		switch state {
		case 1, 2:
			msg.Body = fmt.Sprintf("%s to %f F", state, f)
		case 0:
			msg.Body = "turn off home furnace"
		}

		log.Println("update message body", msg.Body)

		var buf bytes.Buffer
		json.NewEncoder(&buf).Encode(msg)
		resp, err := http.Post(fmt.Sprintf("%s/gadgets", h.furnaceHost), "application/json", &buf)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.Println("unable to update furnace", resp.StatusCode)
		}
	})

	var i int
	h.updateFurnace = func(msg *gogadgets.Message) bool {
		if msg.Sender == "home temperature" {
			f, ok := msg.Value.Value.(float64)
			if ok {
				if i == 0 {
					c := (f - 32.0) / 1.8
					furnace.Thermostat.CurrentTemperature.UpdateValue(c)
				}
				i++
				if i == 10 {
					i = 0
				}
			}
		}
		return false
	}

	h.furnaceOn = func(val *gogadgets.Value) error {
		if val == nil {
			return nil
		}

		if strings.Index(val.Cmd, "heat home") == 0 {
			state = heat
		} else if strings.Index(val.Cmd, "cool home") == 0 {
			state = cool
		}

		furnace.Thermostat.TargetHeatingCoolingState.UpdateValue(int(state))
		furnace.Thermostat.CurrentHeatingCoolingState.UpdateValue(int(state))
		f, ok := val.Value.(float64)
		log.Printf("home kit ON!: %s state: %s, ok: %v", val.Cmd, state, ok)
		if ok {
			c := (f - 32.0) / 1.8
			furnace.Thermostat.TargetTemperature.UpdateValue(c)
		}
		return nil
	}

	h.furnaceOff = func() error {
		state = off
		furnace.Thermostat.TargetHeatingCoolingState.UpdateValue(int(state))
		furnace.Thermostat.CurrentHeatingCoolingState.UpdateValue(int(state))
		return nil
	}

	return furnace.Accessory
}

func (h *Homekit) init() error {
	if h.furnaceHost != "" {
		state, temperature := h.getFurnace()
		if state != off {
			h.furnaceOn(&gogadgets.Value{Cmd: string(state)})
		}

		h.updateFurnace(&gogadgets.Message{Sender: "home temperature", Value: gogadgets.Value{Value: temperature}})
	}
	return nil
}

func (h *Homekit) getFurnace() (homekitState, float64) {
	resp, err := http.Get(fmt.Sprintf("%s/gadgets", h.furnaceHost))
	if err != nil {
		return off, 0.0
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Println("unable to update furnace", resp.StatusCode)
		return off, 0.0
	}

	var state homekitState

	var status map[string]gogadgets.Message
	json.NewDecoder(resp.Body).Decode(&status)
	val := status["home furnace"].Value
	if val.Value == true && strings.Index(val.Cmd, "heat") == 0 {
		state = heat
	} else if val.Value == true && strings.Index(val.Cmd, "cool") == 0 {
		state = cool
	}

	var c float64
	f, ok := status["home temperature"].Value.Value.(float64)
	if ok {
		c = (f - 32.0) / 1.8
	}

	return state, c
}
