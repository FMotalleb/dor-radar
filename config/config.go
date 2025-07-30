package config

import (
	"net/url"
)

type Config struct {
	Core      CoreCfg   `mapstructure:"core"`
	Collector Collector `mapstructure:"collector"`
}

type CoreCfg struct {
	Listen string `mapstructure:"listen"`
}

type Collector struct {
	Target  *url.URL      `mapstructure:"target,omitempty"`
	Filter  string        `mapstructure:"filter,omitempty"`
	Reshape []ReshapeRule `mapstructure:"reshape,omitempty"`
	// Timeout time.Duration `mapstructure:"timeout,omitempty" toml:"timeout,omitempty" yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

type ReshapeRule struct {
	From  string   `mapstructure:"from,omitempty"`
	To    string   `mapstructure:"to,omitempty"`
	Attrs []string `mapstructure:"attrs,omitempty"`
	Size  int      `mapstructure:"size,omitempty"`
}

func (c *Collector) GetShape(key string) string {
	for _, r := range c.Reshape {
		if r.From == key {
			return r.To
		}
	}
	return key
}

func (c *Collector) GetAttrs(key string) []string {
	for _, r := range c.Reshape {
		if r.From != key {
			continue
		}
		if r.Attrs != nil {
			return r.Attrs
		}
		break
	}
	return make([]string, 0)
}

func (c *Collector) GetSize(key string) int {
	for _, r := range c.Reshape {
		if r.From == key {
			return r.Size
		}
	}
	return 15
}
