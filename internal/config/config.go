package config

import (
	"fmt"
	"strings"
)

type Configer interface {
	GetConf() (*Config, error)
}

type Config struct {
	App     *AppConf     `yaml:"app"`
	Command *CommandConf `yaml:"command"`
}

type AppConf struct {
	Listen     string `yaml:"listen"`
	Auth       string `yaml:"auth"`
	DataDir    string `yaml:"data_dir"`
	DisplayUrl string `yaml:"display_url"`
}

type CommandConf struct {
	Whitelist []string `yaml:"whitelist"`
}

func (c *Config) Validate() error {
	if c.App.Listen == "" {
		c.App.Listen = "0.0.0.0:3000"
	}
	if c.App.Auth == "" {
		return fmt.Errorf("app.auth required")
	}
	if c.App.DataDir == "" {
		return fmt.Errorf("app.data_dir required")
	}

	if !strings.HasPrefix(c.App.DisplayUrl, "https://") && !strings.HasPrefix(c.App.DisplayUrl, "http://") {
		return fmt.Errorf("app.display_url required https:// or http:// begin")
	}

	return nil
}
