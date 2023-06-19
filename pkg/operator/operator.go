package operator

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type Command struct {
	ctx     context.Context
	cmd     string
	timeout time.Duration
	diag    diag.Diagnostics
}

func NewCommand(ctx context.Context, cmd string) Command {
	dur, _ := time.ParseDuration("6s")
	return Command{
		ctx:     ctx,
		cmd:     cmd,
		timeout: dur,
	}
}

// CommandOperator executes a command on a machine to install k3sup
type CommandOperator interface {
	Execute(command string) (CommandRes, error)
	ExecuteStdout(command string, stream bool) (CommandRes, error)
}

// CommandRes contains the STDIO output from running a command
type CommandRes struct {
	StdOut   []byte
	StdErr   []byte
	ExitCode int
}
