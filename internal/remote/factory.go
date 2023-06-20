package remote

type ProvisionerFactory struct {
	config *Config
}

func NewFactory(config *Config) *ProvisionerFactory {
	return &ProvisionerFactory{
		config: config,
	}
}

func (f *ProvisionerFactory) Create(config *Config) (*Provisioner, error) {
	if config == nil {
		return NewProvisioner(f.config)
	}
	var cfg Config
	if config.Host != "" {
		cfg.Host = config.Host
	} else {
		cfg.Host = f.config.Host
	}

	if config.Port != "" {
		cfg.Port = config.Port
	} else {
		cfg.Port = f.config.Port
	}

	if config.User != "" {
		cfg.User = config.User
	} else {
		cfg.User = f.config.User
	}

	if config.Password != "" {
		cfg.Password = config.Password
	} else {
		cfg.Password = f.config.Password
	}

	if config.PrivateKey != "" {
		cfg.PrivateKey = config.PrivateKey
	} else {
		cfg.PrivateKey = f.config.PrivateKey
	}

	if config.PrivateKeyPath != "" {
		cfg.PrivateKeyPath = config.PrivateKeyPath
	} else {
		cfg.PrivateKeyPath = f.config.PrivateKeyPath
	}

	if config.RetryDelay != 0 {
		cfg.RetryDelay = config.RetryDelay
	} else {
		cfg.RetryDelay = f.config.RetryDelay
	}

	if config.Timeout != 0 {
		cfg.Timeout = config.Timeout
	} else {
		cfg.Timeout = f.config.Timeout
	}

	return NewProvisioner(&cfg)
}
