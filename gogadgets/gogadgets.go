package gogadgets

import (
	"crypto/rand"
	"fmt"
	"io"
	"sync"
	"time"
)

var (
	COMMAND      = "command"
	METHOD       = "method"
	DONE         = "done"
	UPDATE       = "update"
	GADGET       = "gadget"
	STATUS       = "status"
	METHODUPDATE = "method update"
)

type Value struct {
	Value  interface{} `json:"value,omitempty"`
	Units  string      `json:"units,omitempty"`
	Output interface{} `json:"io,omitempty"`
	ID     string      `json:"id,omitempty"`
}

func (v *Value) ToFloat() (f float64, ok bool) {
	switch V := v.Value.(type) {
	case bool:
		if V {
			f = 1.0
		} else {
			f = 0.0
		}
		ok = true
	case float64:
		f = V
		ok = true
	}
	return f, ok
}

type Message struct {
	UUID        string    `json:"uuid"`
	From        string    `json:"from,omitempty"`
	Name        string    `json:"name,omitempty"`
	Location    string    `json:"location,omitempty"`
	Type        string    `json:"type,omitempty"`
	Sender      string    `json:"sender,omitempty"`
	Target      string    `json:"target,omitempty"`
	Body        string    `json:"body,omitempty"`
	Host        string    `json:"host,omitempty"`
	Method      Method    `json:"method,omitempty"`
	Timestamp   time.Time `json:"timestamp,omitempty"`
	Value       Value     `json:"value,omitempty"`
	TargetValue *Value    `json:"targetValue,omitempty"`
	Info        Info      `json:"info,omitempty"`
	Config      Config    `json:"config,omitempty"`
}

type Method struct {
	Step  int      `json:"step,omitempty"`
	Steps []string `json:"steps,omitempty"`
	Time  int      `json:"time,omitempty"`
}

type Info struct {
	Direction string `json:"direction,omitempty"`
	On        string `json:"on,omitempty"`
	Off       string `json:"off,omitempty"`
}

type GadgetConfig struct {
	Type         string                 `json:"type,omitempty"`
	Location     string                 `json:"location,omitempty"`
	Name         string                 `json:"name,omitempty"`
	OnCommand    string                 `json:"onCommand,omitempty"`
	OffCommand   string                 `json:"offCommand,omitempty"`
	InitialValue string                 `json:"initialValue,omitempty"`
	Pin          Pin                    `json:"pin,omitempty"`
	Args         map[string]interface{} `json:"args,omitempty"`
}

type Config struct {
	Master  string         `json:"master,omitempty"`
	Host    string         `json:"host,omitempty"`
	Port    int            `json:"port,omitempty"`
	Gadgets []GadgetConfig `json:"gadgets,omitempty"`
	Logger  Logger         `json:"-"`
}

type Logger interface {
	Println(...interface{})
	Printf(string, ...interface{})
	Fatal(...interface{})
}

type Pin struct {
	Type        string                 `json:"type,omitempty"`
	Port        string                 `json:"port,omitempty"`
	Pin         string                 `json:"pin,omitempty"`
	Direction   string                 `json:"direction,omitempty"`
	Edge        string                 `json:"edge,omitempty"`
	ActiveLow   string                 `json:"active_low,omitempty"`
	OneWirePath string                 `json:"onewirePath,omitempty"`
	OneWireId   string                 `json:"onewireId,omitempty"`
	Sleep       time.Duration          `json:"sleep,omitempty"`
	Value       interface{}            `json:"value,omitempty"`
	Units       string                 `json:"units,omitempty"`
	Platform    string                 `json:"platform,omitempty"`
	Frequency   int                    `json:"frequency,omitempty"`
	Args        map[string]interface{} `json:"args,omitempty"`
	Pins        map[string]Pin         `json:"pins,omitempty"`
	Lock        sync.Mutex             `json:"-"`
}

func GetUUID() string {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return ""
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}
