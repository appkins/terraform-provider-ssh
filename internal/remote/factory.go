package remote

import (
	"context"

	"github.com/appkins/terraform-provider-ssh/internal/ssh"
)

type ProvisionerFactory struct {
	config *ssh.Config
}

func NewFactory(config *ssh.Config) *ProvisionerFactory {
	return &ProvisionerFactory{
		config: config,
	}
}

func (f *ProvisionerFactory) Create(ctx context.Context, config *ssh.Config) (*Provisioner, error) {
	if config == nil {
		return createProvisioner(ctx, f.config)
	}
	var cfg ssh.Config
	if config.Host.ValueString() != "" {
		cfg.Host = config.Host
	} else {
		cfg.Host = f.config.Host
	}

	if config.Port.ValueInt64() != 0 {
		cfg.Port = config.Port
	} else {
		cfg.Port = f.config.Port
	}

	if config.User.ValueString() != "" {
		cfg.User = config.User
	} else {
		cfg.User = f.config.User
	}

	if config.Password.ValueString() != "" {
		cfg.Password = config.Password
	} else {
		cfg.Password = f.config.Password
	}

	if config.PrivateKey.ValueString() != "" {
		cfg.PrivateKey = config.PrivateKey
	} else {
		cfg.PrivateKey = f.config.PrivateKey
	}

	if config.PrivateKeyPath.ValueString() != "" {
		cfg.PrivateKeyPath = config.PrivateKeyPath
	} else {
		cfg.PrivateKeyPath = f.config.PrivateKeyPath
	}

	if config.RetryDelay.ValueInt64() != 0 {
		cfg.RetryDelay = config.RetryDelay
	} else {
		cfg.RetryDelay = f.config.RetryDelay
	}

	if config.Timeout.ValueInt64() != 0 {
		cfg.Timeout = config.Timeout
	} else {
		cfg.Timeout = f.config.Timeout
	}

	return createProvisioner(ctx, &cfg)
}
