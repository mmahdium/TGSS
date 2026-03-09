package config

import "time"

const (
	DialTimeout         = 20 * time.Second
	ExchangeTimeout     = 20 * time.Second
	MigrationTimeout    = 20 * time.Second
	AuthOperationTimeout    = 20 * time.Second
	MessageOperationTimeout = 10 * time.Second
	AuthStatusTimeout       = 10 * time.Second
)
