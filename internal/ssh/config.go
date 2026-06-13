package ssh

import (
	"fmt"
	"os"
	"time"

	"github.com/appleboy/easyssh-proxy"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/crypto/ssh"
)

type Config struct {
	Host           types.String `tfsdk:"host"`
	Port           types.Int64  `tfsdk:"port"`
	User           types.String `tfsdk:"user"`
	Password       types.String `tfsdk:"password"`
	PrivateKey     types.String `tfsdk:"private_key"`
	PrivateKeyPath types.String `tfsdk:"private_key_path"`
	Timeout        types.Int64  `tfsdk:"timeout"`
	RetryDelay     types.Int64  `tfsdk:"retry_delay"`
}

func (c *Config) getClientAuth() (methods []ssh.AuthMethod, err error) {
	if p := c.PrivateKeyPath.ValueString(); p != "" {
		if b, err := os.ReadFile(p); err == nil {
			if sig, err := ssh.ParsePrivateKey(b); err == nil {
				methods = append(methods, ssh.PublicKeys(sig))
			}
		}
	}
	if key := c.PrivateKey.ValueString(); key != "" {
		if sig, err := ssh.ParsePrivateKey([]byte(key)); err == nil {
			methods = append(methods, ssh.PublicKeys(sig))
		}
	}
	if pw := c.Password.ValueString(); pw != "" {
		methods = append(methods, ssh.Password(pw))
	}
	return
}

func (c *Config) getEasyConfig() *easyssh.MakeConfig {
	user := c.User.ValueString()
	if user == "" {
		user = "root"
	}
	port := c.Port.ValueInt64()
	if port == 0 {
		port = 22
	}
	timeout := c.Timeout.ValueInt64()
	if timeout == 0 {
		timeout = 60
	}

	cfg := easyssh.MakeConfig{
		User:    c.User.ValueString(),
		Server:  c.Host.ValueString(),
		Port:    fmt.Sprintf("%d", port),
		Timeout: time.Duration(timeout),
	}

	if pass := c.Password.ValueString(); pass != "" {
		cfg.Password = pass
	}
	if key := c.PrivateKey.ValueString(); key != "" {
		cfg.Key = key
	}
	if keyPath := c.PrivateKeyPath.ValueString(); keyPath != "" {
		cfg.KeyPath = keyPath
	}

	return &cfg
}

func (c *Config) getClientConfig() (*ssh.ClientConfig, error) {
	auth, err := c.getClientAuth()
	if err != nil {
		return &ssh.ClientConfig{}, err
	}

	return &ssh.ClientConfig{
		User:            c.User.ValueString(),
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            auth,
	}, nil
}
