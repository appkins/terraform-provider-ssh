package provider

import (
	"fmt"
	"time"

	"github.com/appkins/terraform-provider-ssh/internal/remote"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SshConfig struct {
	Host           types.String `tfsdk:"host"`
	Port           types.String `tfsdk:"port"`
	User           types.String `tfsdk:"user"`
	Password       types.String `tfsdk:"password"`
	PrivateKey     types.String `tfsdk:"private_key"`
	PrivateKeyPath types.String `tfsdk:"private_key_path"`
	Timeout        types.Int64  `tfsdk:"timeout"`
	RetryDelay     types.Int64  `tfsdk:"retry_delay"`
}

func GetSshConfig(config *SshConfig) *remote.Config {

	if config == nil {
		return &remote.Config{}
	}

	cfg := &remote.Config{
		Host:           config.Host.ValueString(),
		Port:           config.Port.ValueString(),
		User:           config.User.ValueString(),
		Password:       config.Password.ValueString(),
		PrivateKey:     config.PrivateKey.ValueString(),
		PrivateKeyPath: config.PrivateKeyPath.ValueString(),
	}

	if !config.Timeout.IsUnknown() && !config.Timeout.IsNull() {
		if t, err := time.ParseDuration(fmt.Sprintf("%ds", config.Timeout.ValueInt64())); err == nil {
			cfg.Timeout = t
		}
	} else {
		cfg.Timeout = 30 * time.Second
	}

	if !config.RetryDelay.IsUnknown() && !config.RetryDelay.IsNull() {
		if t, err := time.ParseDuration(fmt.Sprintf("%ds", config.RetryDelay.ValueInt64())); err == nil {
			cfg.Timeout = t
		}
	} else {
		cfg.Timeout = 30 * time.Second
	}

	return cfg
}
