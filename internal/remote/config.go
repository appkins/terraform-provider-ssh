package remote

import (
	"time"

	"github.com/loafoe/easyssh-proxy/v2"
)

type Config struct {
	Host           string
	Port           string
	User           string
	Password       string
	PrivateKey     string
	PrivateKeyPath string
	RetryDelay     time.Duration
	Timeout        time.Duration
}

func (p *Provisioner) CreateClient(config *Config) (*easyssh.MakeConfig, error) {
	cfg := p.MakeConfig
	if config == nil {
		return cfg, nil
	}
	if config.Host != "" {
		cfg.Server = config.Host
	}

	if config.Port != "" {
		cfg.Port = config.Port
	}

	if config.User != "" {
		cfg.User = config.User
	}

	if config.Password != "" {
		cfg.Password = config.Password
	}

	if config.PrivateKey != "" {
		cfg.Key = config.PrivateKey
	}

	if config.PrivateKeyPath != "" {
		cfg.KeyPath = config.PrivateKeyPath
	}

	if config.Timeout != 0 {
		cfg.Timeout = config.Timeout
	}

	return cfg, nil

}
