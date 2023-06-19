package remote

import (
	"context"
	"time"

	"github.com/loafoe/easyssh-proxy/v2"
)

type Provisioner struct {
	Ssh        *easyssh.MakeConfig
	Timeout    time.Duration
	RetryDelay time.Duration
}

func (p *Provisioner) Execute(commands []string, ctx context.Context) (string, error) {
	return exec(ctx, p.RetryDelay, commands, p.Timeout, p.Ssh)
}

func (p *Provisioner) CopyFiles(files []File, ctx context.Context) error {
	return copyFiles(ctx, p.RetryDelay, p.Ssh, files)
}

func NewProvisioner(ssh *easyssh.MakeConfig, timeout time.Duration, retryDelay time.Duration) *Provisioner {
	return &Provisioner{
		Ssh:        ssh,
		Timeout:    timeout,
		RetryDelay: retryDelay,
	}
}
