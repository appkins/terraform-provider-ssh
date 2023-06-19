// The SSHOperator struct executes commands on a remote machine over an SSH session.
package operator

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/yahoo/vssh"
	"golang.org/x/crypto/ssh"
)

// SSHOperator executes commands on a remote machine over an SSH session
type SSHOperator struct {
	*vssh.VSSH
}

func NewSSHOperator() *SSHOperator {
	operator := &SSHOperator{VSSH: vssh.New()}
	operator.Start()
	return operator
}

func (s *SSHOperator) AddClient(address string, config *ssh.ClientConfig) (*SSHOperator, error) {
	if err := s.VSSH.AddClient(address, config); err != nil {
		return s, err
	}
	s.Wait()
	return s, nil
}

func (s *SSHOperator) CopyFile(ctx context.Context, source string, dest string) error {
	// Open the source file
	srcFile, err := os.Open(source)
	if err != nil {
		panic(err)
	}

	// Get the file size
	stat, err := srcFile.Stat()
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	timeout, _ := time.ParseDuration("20s")

	respChan := s.Run(ctx, fmt.Sprintf("scp -s %s", dest), timeout)

	resp := <-respChan
	if err := resp.Err(); err != nil {
		log.Fatal(err)
	}

	ss := resp.GetStream()
	defer s.Close()

	buf := new(bytes.Buffer)

	ss.Input(buf)

	fmt.Fprintf(buf, "C0644 %d %s\n", stat.Size(), path.Base(source))

	io.Copy(buf, srcFile)
	fmt.Fprint(buf, "\x00")

	for ss.ScanStdout() {
		txt := ss.TextStdout()
		fmt.Println(txt)
	}

	return nil
}

func (s *SSHOperator) ExecuteStdout(command Command, stream bool) (CommandRes, error) {

	ctx, cancel := context.WithCancel(command.ctx)

	defer cancel()

	respChan := s.Run(ctx, command.cmd, command.timeout)

	for resp := range respChan {

		if stream {
			resp := <-respChan
			if err := resp.Err(); err != nil {
				log.Fatal(err)
			}

			s := resp.GetStream()
			defer s.Close()

			output, errorOutput := new(strings.Builder), new(strings.Builder)

			for s.ScanStdout() {
				output.WriteString(s.TextStdout())
			}

			for s.ScanStderr() {
				errorOutput.WriteString(s.TextStdout())
			}

			return CommandRes{
				StdOut: []byte(output.String()),
				StdErr: []byte(errorOutput.String()),
			}, nil
		} else {

			if err := resp.Err(); err != nil {
				command.diag.AddError("Client Error", fmt.Sprintf("Unable to run command, got error: %s", err))
				continue
			}

			output, errorOutput, _ := resp.GetText(s.VSSH)

			return CommandRes{
				StdOut: []byte(output),
				StdErr: []byte(errorOutput),
			}, nil
		}
	}
	return CommandRes{}, nil
}

func (s *SSHOperator) Execute(command Command) (CommandRes, error) {
	return s.ExecuteStdout(command, true)
}

func (s *SSHOperator) Close() error {
	return nil
}
