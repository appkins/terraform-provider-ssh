package remote

import (
	"context"
	"time"

	"github.com/appkins/terraform-provider-ssh/internal/ssh"
	"github.com/loafoe/easyssh-proxy/v2"
)

type Provisioner struct {
	ssh        *easyssh.MakeConfig
	ctx        context.Context
	Timeout    time.Duration
	RetryDelay time.Duration
}

func createProvisioner(ctx context.Context, config *ssh.Config) (*Provisioner, error) {
	return &Provisioner{
		ssh: &easyssh.MakeConfig{
			User:     config.User.ValueString(),
			Server:   config.Host.ValueString(),
			Port:     string(config.Port.ValueInt64()),
			Timeout:  time.Duration(config.Timeout.ValueInt64()),
			Password: config.Password.ValueString(),
			Key:      config.PrivateKey.ValueString(),
			KeyPath:  config.PrivateKeyPath.ValueString(),
		},
		ctx:        ctx,
		Timeout:    time.Duration(config.Timeout.ValueInt64()),
		RetryDelay: time.Duration(config.RetryDelay.ValueInt64()),
	}, nil
}
