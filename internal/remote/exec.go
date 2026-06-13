package remote

import (
	"fmt"
	"strings"
	"time"

	"github.com/appkins/terraform-provider-ssh/internal/log"
)

func (p *Provisioner) ExecuteCommand(command string) (output string, err error) {
	var stdout, stderr string
	var done bool
	for {
		stdout, stderr, done, err = p.ssh.Run(command, p.Timeout)
		log.Debug(p.ctx, "command: %s\ndone: %t\nstdout: %s\nstderr: %s\nerror: %s", command, done, stdout, stderr, err)
		if err == nil {
			output = strings.TrimSuffix(stdout, "\n")
			break
		}
		if strings.Contains(err.Error(), "no supported methods remain") {
			return
		}

		select {
		case <-time.After(p.RetryDelay):
		case <-p.ctx.Done():
			log.Debug(p.ctx, "error: %v\n", err)
			log.Error(p.ctx, "execution of command '%s' failed: %s: %s", command, p.ctx.Err(), err)
			if stderr != "" {
				err = fmt.Errorf("stderr output: %s", stderr)
				return
			}
			return
		}
	}
	return
}

/**
 * Provisioner is a struct that contains the SSH client and the retry delay.
 * The SSH client is used to execute commands on the remote host.
 * The retry delay is used to determine how long to wait before retrying a failed command.
 */
func (p *Provisioner) Execute(commands []string) (output []string, err error) {
	for _, command := range commands {
		var stdout, stderr string
		var done bool
		for {
			stdout, stderr, done, err = p.ssh.Run(command, p.Timeout)
			log.Debug(p.ctx, "command: %s\ndone: %t\nstdout: %s\nstderr: %s\nerror: %s", command, done, stdout, stderr, err)
			if err == nil {
				output = append(output, strings.TrimSuffix(stdout, "\n"))
				break
			}
			if strings.Contains(err.Error(), "no supported methods remain") {
				return
			}

			select {
			case <-time.After(p.RetryDelay):
			case <-p.ctx.Done():
				log.Debug(p.ctx, "error: %v\n", err)
				log.Error(p.ctx, "execution of command '%s' failed: %s: %s", command, p.ctx.Err(), err)
				if stderr != "" {
					err = fmt.Errorf("stderr output: %s", stderr)
					return
				}
				return
			}
		}
	}
	return
}
