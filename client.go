package goconsul

import (
	"strconv"
	"sync"

	"github.com/hashicorp/consul/api"
)

type Client struct {
	api.Client
	insMap sync.Map
}

func DefaultClient() *Client {
	return &Client{}
}

func NewClient(host string, port int, token string) (*Client, error) {

	c := &Client{}

	if err := c.Connect(host, port, token); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) Connect(host string, port int, token string) error {
	config := api.DefaultConfig()
	if len(host) > 3 && port > 0 && port <= 65535 {
		config.Address = host + ":" + strconv.Itoa(port)
	}

	config.Token = token
	client, err := api.NewClient(config)
	if err != nil {
		return err
	}

	c.Client = *client

	return nil
}
