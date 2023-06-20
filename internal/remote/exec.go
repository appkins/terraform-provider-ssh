package remote

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/appkins/terraform-provider-ssh/internal/log"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

/**
 * Provisioner is a struct that contains the SSH client and the retry delay.
 * The SSH client is used to execute commands on the remote host.
 * The retry delay is used to determine how long to wait before retrying a failed command.
 */
func (p *Provisioner) Execute(ctx context.Context, commands []string) ([]string, error) {

	var output []string

	var stdout, stderr string
	var done bool
	var err error

	for _, command := range commands {
		for {
			stdout, stderr, done, err = p.ssh.Run(command, p.Timeout)
			log.Debug(ctx, "command: %s\ndone: %t\nstdout: %s\nstderr: %s\nerror: %s", command, done, stdout, stderr, err)
			if err == nil {
				output = append(output, strings.TrimSuffix(stdout, "\n"))
				break
			}
			if strings.Contains(err.Error(), "no supported methods remain") {
				return output, err
			}

			select {
			case <-time.After(p.RetryDelay):
				// Retry

			case <-ctx.Done():
				tflog.Debug(ctx, fmt.Sprintf("error: %v\n", err))
				tflog.Error(ctx, fmt.Sprintf("execution of command '%s' failed: %s: %s", command, ctx.Err(), err))
				if stderr != "" {
					return output, fmt.Errorf("stderr output: %s", stderr)
				}
				return output, err
			}
		}
	}
	return output, nil
}
