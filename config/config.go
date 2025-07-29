package config

import (
	"net/url"
	"time"
)

type Config struct {
	Collectors []Collector `mapstructure:"collectors" toml:"collectors" yaml:"collectors" json:"collectors"`
}

type Collector struct {
	Target  *url.URL      `mapstructure:"to,omitempty" toml:"to,omitempty" yaml:"to,omitempty" json:"to,omitempty"`
	Timeout time.Duration `mapstructure:"timeout,omitempty" toml:"timeout,omitempty" yaml:"timeout,omitempty" json:"timeout,omitempty"`
}
