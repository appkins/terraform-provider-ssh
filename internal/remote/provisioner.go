package remote

import (
	"time"

	"github.com/loafoe/easyssh-proxy/v2"
)

type Provisioner struct {
	*easyssh.MakeConfig
	Timeout    time.Duration
	RetryDelay time.Duration
}

func NewProvisioner(config *Config) *Provisioner {
	return &Provisioner{
		MakeConfig: &easyssh.MakeConfig{
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
	}
}
