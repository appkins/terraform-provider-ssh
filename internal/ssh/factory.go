package ssh

import "github.com/hashicorp/terraform-plugin-framework/types"

type Factory struct {
	defaultConfig Config
}

func NewFactory(defaultConfig Config) *Factory {
	return &Factory{
		defaultConfig: defaultConfig,
	}
}

func (f *Factory) Create(config Config) (*Client, error) {
	res := Config{}
	if !config.Host.IsUnknown() && config.Host.ValueString() != "" {
		res.Host = config.Host
	} else if !f.defaultConfig.Host.IsUnknown() && f.defaultConfig.Host.ValueString() != "" {
		res.Host = f.defaultConfig.Host
	}
	if !config.Port.IsUnknown() && config.Port.ValueInt64() != 0 {
		res.Port = config.Port
	} else if !f.defaultConfig.Port.IsUnknown() && f.defaultConfig.Port.ValueInt64() != 0 {
		res.Port = f.defaultConfig.Port
	} else {
		res.Port = types.Int64Value(22)
	}
	if config.User.ValueString() != "" {
		res.User = config.User
	} else if f.defaultConfig.User.ValueString() != "" {
		res.User = f.defaultConfig.User
	} else {
		res.User = types.StringValue("root")
	}
	if config.Password.ValueString() != "" {
		res.Password = config.Password
	} else if f.defaultConfig.Password.ValueString() != "" {
		res.Password = f.defaultConfig.Password
	}
	if config.PrivateKey.ValueString() != "" {
		res.PrivateKey = config.PrivateKey
	} else if f.defaultConfig.PrivateKey.ValueString() != "" {
		res.PrivateKey = f.defaultConfig.PrivateKey
	}
	if config.PrivateKeyPath.ValueString() != "" {
		res.PrivateKeyPath = config.PrivateKeyPath
	} else if f.defaultConfig.PrivateKeyPath.ValueString() != "" {
		res.PrivateKeyPath = f.defaultConfig.PrivateKeyPath
	}
	if config.RetryDelay.ValueInt64() != 0 {
		res.RetryDelay = config.RetryDelay
	} else if f.defaultConfig.RetryDelay.ValueInt64() != 0 {
		res.RetryDelay = f.defaultConfig.RetryDelay
	}
	if config.Timeout.ValueInt64() != 0 {
		res.Timeout = config.Timeout
	} else if f.defaultConfig.Timeout.ValueInt64() != 0 {
		res.Timeout = f.defaultConfig.Timeout
	}

	return NewClient(&config)

}
