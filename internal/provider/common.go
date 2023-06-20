package provider

import (
	"fmt"
	"time"

	"github.com/appkins/terraform-provider-ssh/internal/remote"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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

var SshDatasourceBlock = map[string]dschema.Block{
	"ssh": dschema.SingleNestedBlock{
		Attributes: map[string]dschema.Attribute{
			"host": dschema.StringAttribute{
				MarkdownDescription: "SSH host",
				Optional:            true,
			},
			"port": dschema.StringAttribute{
				MarkdownDescription: "SSH port",
				Optional:            true,
			},
			"user": dschema.StringAttribute{
				MarkdownDescription: "SSH user",
				Optional:            true,
			},
			"password": dschema.StringAttribute{
				MarkdownDescription: "SSH password",
				Optional:            true,
				Sensitive:           true,
			},
			"private_key": dschema.StringAttribute{
				MarkdownDescription: "SSH private key data",
				Optional:            true,
				Sensitive:           true,
			},
			"private_key_path": dschema.StringAttribute{
				MarkdownDescription: "Path to SSH private key",
				Optional:            true,
			},
			"timeout": dschema.StringAttribute{
				MarkdownDescription: "Timeout",
				Optional:            true,
			},
			"retry_delay": dschema.StringAttribute{
				MarkdownDescription: "Retry delay",
				Optional:            true,
			},
		},
	},
}

func GetSshConfig(config SshConfig) *remote.Config {
	cfg := &remote.Config{
		Host:       config.Host.ValueString(),
		Port:       config.Port.ValueString(),
		User:       config.User.ValueString(),
		Password:   config.Password.ValueString(),
		PrivateKey: config.PrivateKey.ValueString(),
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
