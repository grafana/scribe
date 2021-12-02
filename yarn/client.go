package yarn

type Client struct {
}

func (c *Client) Install() func() error {
	return NewStep("install")
}
