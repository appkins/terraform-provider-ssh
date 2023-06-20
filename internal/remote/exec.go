package remote

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (p *Provisioner) Execute(ctx context.Context, commands []string, config *Config) (string, error) {

	ssh, _ := p.CreateClient(config)

	retryDelay := p.RetryDelay

	if config.RetryDelay != 0 {
		retryDelay = config.RetryDelay
	}

	var stdout, stderr string
	var done bool
	var err error

	for i := 0; i < len(commands); i++ {
		for {
			stdout, stderr, done, err = ssh.Run(commands[i], config.Timeout)
			tflog.Debug(ctx, commands[i], map[string]interface{}{"done": done, "stdout": stdout, "stderr": stderr, "error": err})
			if err == nil {
				break
			}
			if strings.Contains(err.Error(), "no supported methods remain") {
				return stdout, err
			}

			select {
			case <-time.After(retryDelay):
				// Retry

			case <-ctx.Done():
				tflog.Debug(ctx, fmt.Sprintf("error: %v\n", err))
				tflog.Error(ctx, fmt.Sprintf("execution of command '%s' failed: %s: %s", commands[i], ctx.Err(), err))
				if stderr != "" {
					return stdout, fmt.Errorf("stderr output: %s", stderr)
				}
				return stdout, err
			}
		}
	}
	return stdout, nil
}
