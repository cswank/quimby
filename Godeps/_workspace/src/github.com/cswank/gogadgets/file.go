package gogadgets

import (
	"bytes"
	"io/ioutil"
)

// File is a way to a debug gadgets system.  Doesn't
// really do anything.
type File struct {
	path string
}

func NewFile(pin *Pin) (OutputDevice, error) {
	return &File{
		path: pin.Args["path"].(string),
	}, nil
}

func (f *File) Config() ConfigHelper {
	return ConfigHelper{}
}

func (f *File) Update(msg *Message) {

}

func (f *File) On(val *Value) error {
	return ioutil.WriteFile(f.path, []byte("1"), 0666)
}

func (f *File) Status() interface{} {
	data, err := ioutil.ReadFile(f.path)
	return err == nil && bytes.Equal(data, []byte("1"))
}

func (f *File) Off() error {
	return ioutil.WriteFile(f.path, []byte("0"), 0666)
}