package config

import (
	"net/url"
)

type Config struct {
	Core CoreCfg `mapstructure:"core" toml:"core" yaml:"core" json:"core"`
}

type CoreCfg struct {
	Listen    string    `mapstructure:"listen" toml:"listen" yaml:"listen" json:"listen"`
	Collector Collector `mapstructure:"collector" toml:"collector" yaml:"collector" json:"collector"`
}

type Collector struct {
	Target *url.URL `mapstructure:"target,omitempty" toml:"target,omitempty" yaml:"target,omitempty" json:"target,omitempty"`
	// Timeout time.Duration `mapstructure:"timeout,omitempty" toml:"timeout,omitempty" yaml:"timeout,omitempty" json:"timeout,omitempty"`
}
