package quimby

//Option is used to intialize a Client
type Option func(c *Client)

//Token is used to apply a jwt token to the client so you
//don't have to authenticate every time.
func Token(t string) Option {
	return func(c *Client) {
		c.token = t
	}
}
