package ssh

import (
	"github.com/appleboy/easyssh-proxy"
)

type Client struct {
	*easyssh.MakeConfig
}

func NewClient(c *Config) (*Client, error) {
	return &Client{c.getEasyConfig()}, nil
}
