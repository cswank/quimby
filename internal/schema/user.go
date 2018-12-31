package schema

// User represents a user
type User struct {
	ID       int `storm:"id,increment"`
	Name     string
	Password []byte
}
