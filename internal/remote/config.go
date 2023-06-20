package remote

import (
	"time"
)

type Config struct {
	Host           string
	Port           string
	User           string
	Password       string
	PrivateKey     string
	PrivateKeyPath string
	RetryDelay     time.Duration
	Timeout        time.Duration
}
