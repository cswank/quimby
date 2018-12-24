package schema

// Gadget represents a gadget
type Gadget struct {
	ID   int    `storm:"id,increment"`
	Name string `json:"name"`
	URL  string `json:"url"`
}
