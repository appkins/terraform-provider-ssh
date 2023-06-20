package remote

import (
	"time"

	"github.com/loafoe/easyssh-proxy/v2"
)

type Provisioner struct {
	ssh        *easyssh.MakeConfig
	Timeout    time.Duration
	RetryDelay time.Duration
}

func NewProvisioner(config *Config) (*Provisioner, error) {
	return &Provisioner{
		ssh: &easyssh.MakeConfig{
			User:     config.User,
			Server:   config.Host,
			Port:     config.Port,
			Timeout:  config.Timeout,
			Password: config.Password,
			Key:      config.PrivateKey,
			KeyPath:  config.PrivateKeyPath,
		},
		Timeout:    config.Timeout,
		RetryDelay: config.RetryDelay,
	}, nil
}
