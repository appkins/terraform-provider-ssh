package ssh

import (
	"os"

	"golang.org/x/crypto/ssh"
)

func (c *Config) getAuth() (methods []ssh.AuthMethod, err error) {
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
