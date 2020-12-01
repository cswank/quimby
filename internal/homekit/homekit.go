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
	"github.com/cswank/quimby/internal/config"
	"github.com/kelseyhightower/envconfig"
)

const (
	thermostatOff thermostatState = 0
	heat          thermostatState = 1
	cool          thermostatState = 2
	auto          thermostatState = 3

	on  state = true
	off state = false
)

type (
	state           bool
	thermostatState uint8

	cfg struct {
		Pin            string   `envconfig:"PIN" required:"true"`
		Port           string   `envconfig:"PORT" required:"true"`
		FurnaceHost    string   `envconfig:"FURNACE_HOST"`
		Thermostat     string   `envconfig:"THERMOSTAT" default:"home furnace"`
		Thermometer    string   `envconfig:"THERMOMETER" default:"home temperature"`
		SprinklerHost  string   `envconfig:"SPRINKLER_HOST"`
		SprinklerZones []string `envconfig:"SPRINKLER_ZONES"`
	}

	update func(gogadgets.Message)

	Homekit struct {
		cfg     cfg
		updates map[string]update
	}
)

func (h state) String() string {
	if h {
		return "turn on"
	}

	return "turn off"
}

func (h thermostatState) String() string {
	switch h {
	case 1:
		return "heat home"
	case 2:
		return "cool home"
	default:
		return "turn off home furnace"
	}
}

func New() (*Homekit, error) {
	var c cfg
	err := envconfig.Process("HOMEKIT", &c)
	if err != nil {
		return nil, err
	}

	f := func(msg gogadgets.Message) {
		log.Println("not implemented")
	}

	ss := make(map[string]update, len(c.SprinklerZones))
	for _, z := range c.SprinklerZones {
		ss[z] = f
	}

	h := &Homekit{
		cfg:     c,
		updates: map[string]update{},
	}

	err = h.init()
	if err != nil {
		return nil, err
	}

	go h.start()

	return h, nil
}

func (h *Homekit) Update(msg gogadgets.Message) {
	f, ok := h.updates[msg.Sender]
	if ok {
		f(msg)
	}
}

func (h *Homekit) start() {
	bridge := accessory.NewBridge(accessory.Info{Name: "Quimby"})

	var ac []*accessory.Accessory
	if h.cfg.FurnaceHost != "" {
		ac = append(ac, h.furnace())
	}

	if h.cfg.SprinklerHost != "" {
		ac = append(ac, h.sprinklers()...)
	}

	tr, err := hc.NewIPTransport(
		hc.Config{Pin: h.cfg.Pin, Port: h.cfg.Port},
		bridge.Accessory,
		ac...,
	)

	if err != nil {
		log.Panic(err)
	}

	tr.Start()
}

func (h *Homekit) sprinklers() []*accessory.Accessory {
	out := make([]*accessory.Accessory, len(h.cfg.SprinklerZones))
	m := make(map[string]accessory.Switch)

	for i, z := range h.cfg.SprinklerZones {
		s := accessory.NewSwitch(accessory.Info{Name: z})
		m[z] = *s

		s.Switch.On.OnValueRemoteUpdate(func(b bool) {
			h.sendOnOffCommand(s.Accessory.Info.Name.String.GetValue(), state(b))
		})

		h.updates[z] = func(msg gogadgets.Message) {
			sw, ok := m[msg.Sender]
			b, ok := msg.Value.Value.(bool)
			if ok {
				sw.Switch.On.SetValue(b)
			}
		}

		out[i] = s.Accessory
	}

	return out
}

func (h *Homekit) sendOnOffCommand(name string, val state) {
	msg := gogadgets.Message{Type: gogadgets.COMMAND, Sender: "homekit", Body: fmt.Sprintf("%s %s", val, name)}
	h.sendCommand(msg, h.cfg.SprinklerHost)
}

func (h *Homekit) furnace() *accessory.Accessory {
	furnace := accessory.NewThermostat(accessory.Info{Name: "Thermostat"}, 20, 16, 26, 1)
	state := thermostatOff

	furnace.Thermostat.TargetHeatingCoolingState.OnValueRemoteUpdate(func(i int) {
		state = thermostatState(i) //TODO: figure out how to handle 'auto' state
		c := furnace.Thermostat.TargetTemperature.GetValue()
		h.updateFurnace(c, state)
	})

	furnace.Thermostat.TargetTemperature.OnValueRemoteUpdate(func(c float64) {
		h.updateFurnace(c, state)
	})

	var i int
	h.updates[h.cfg.Thermometer] = func(msg gogadgets.Message) {
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

	h.updates[h.cfg.Thermostat] = func(msg gogadgets.Message) {
		if msg.TargetValue == nil {
			return
		}

		val := *msg.TargetValue
		if strings.Index(val.Cmd, "heat home") == 0 {
			state = heat
		} else if strings.Index(val.Cmd, "cool home") == 0 {
			state = cool
		} else {
			state = thermostatOff
		}

		furnace.Thermostat.TargetHeatingCoolingState.UpdateValue(int(state))
		furnace.Thermostat.CurrentHeatingCoolingState.UpdateValue(int(state))
		if state != thermostatOff {
			f, ok := val.Value.(float64)
			if ok {
				c := (f - 32.0) / 1.8
				furnace.Thermostat.TargetTemperature.UpdateValue(c)
			}
		}
	}

	return furnace.Accessory
}

func (h *Homekit) updateFurnace(c float64, state thermostatState) {
	f := float64(c*1.8 + 32.0)
	msg := gogadgets.Message{Type: gogadgets.COMMAND, Sender: "homekit"}

	switch state {
	case heat, cool:
		msg.Body = fmt.Sprintf("%s to %f F", state, f)
	case thermostatOff:
		msg.Body = "turn off furnace"
	}

	h.sendCommand(msg, h.cfg.FurnaceHost)
}

func (h *Homekit) sendCommand(msg gogadgets.Message, host string) {
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(msg)
	resp, err := http.Post(fmt.Sprintf("%s/gadgets", host), "application/json", &buf)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("unable to update %s: %d", host, resp.StatusCode)
	}
}

func (h *Homekit) init() error {
	if h.cfg.FurnaceHost != "" {
		err := h.register(h.cfg.FurnaceHost)
		if err != nil {
			return err
		}
	}

	if h.cfg.SprinklerHost != "" {
		err := h.register(h.cfg.SprinklerHost)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *Homekit) register(addr string) error {
	cfg := config.Get()

	m := map[string]string{"address": cfg.InternalAddress, "token": "n/a"}

	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(&m)
	if err != nil {
		return err
	}

	r, err := http.Post(fmt.Sprintf("%s/clients", addr), "application/json", buf)
	if err != nil {
		return err
	}

	r.Body.Close()
	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response from %s: %d", addr, r.StatusCode)
	}

	return nil
}
