// The SSHOperator struct executes commands on a remote machine over an SSH session.
package operator

import (
	"context"
	"reflect"
	"testing"

	"github.com/yahoo/vssh"
)

func TestSSHOperator_CopyFile(t *testing.T) {
	type fields struct {
		VSSH *vssh.VSSH
	}
	type args struct {
		ctx    context.Context
		source string
		dest   string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "testfile", fields: fields{vssh.New()}, args: args{context.Background(), "testfile", "/tmp/testfile"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SSHOperator{
				VSSH: tt.fields.VSSH,
			}
			s.Start()
			cfg, _ := vssh.GetConfigPEM("root", "~/.ssh/id_ed25519")
			s.AddClient("192.168.1.198", cfg)
			s.Wait()

			if err := s.CopyFile(tt.args.ctx, tt.args.source, tt.args.dest); (err != nil) != tt.wantErr {
				t.Errorf("SSHOperator.CopyFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSSHOperator_ExecuteStdout(t *testing.T) {
	type fields struct {
		VSSH *vssh.VSSH
	}
	type args struct {
		command Command
		stream  bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    CommandRes
		wantErr bool
	}{
		{name: "ping", fields: fields{vssh.New()}, args: args{NewCommand(context.Background(), "ping 1.1.1.1"), true}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SSHOperator{
				VSSH: tt.fields.VSSH,
			}
			s.Start()
			cfg, _ := vssh.GetConfigPEM("root", "~/.ssh/id_ed25519")
			s.AddClient("192.168.1.198", cfg)
			s.Wait()
			got, err := s.ExecuteStdout(tt.args.command, tt.args.stream)
			if (err != nil) != tt.wantErr {
				t.Errorf("SSHOperator.ExecuteStdout() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SSHOperator.ExecuteStdout() = %v, want %v", got, tt.want)
			}
		})
	}
}
