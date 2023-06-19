package remote

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/loafoe/easyssh-proxy/v2"
)

func exec(ctx context.Context, retryDelay time.Duration, commands []string, timeout time.Duration, ssh *easyssh.MakeConfig) (string, error) {
	var stdout, stderr string
	var done bool
	var err error

	for i := 0; i < len(commands); i++ {
		for {
			stdout, stderr, done, err = ssh.Run(commands[i], timeout)
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
