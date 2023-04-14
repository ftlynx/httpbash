package file

import (
	"fmt"
	"github.com/ftlynx/httpbash/internal/config"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"path/filepath"
	"sync"
)

var (
	cfg  *config.Config
	once sync.Once
)

func NewFileConf(filePath string) config.Configer {
	return &fileConfig{filePath: filePath}
}

type fileConfig struct {
	filePath string
}

func (f *fileConfig) GetConf() (*config.Config, error) {
	var err error

	once.Do(func() {
		err = parseConfig(f.filePath)
	})

	if err != nil {
		return nil, err
	}

	if err = cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func parseConfig(path string) error {
	configPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("get config file absolute path failed, %s", err.Error())
	}

	file, err := os.Open(configPath)
	defer file.Close()
	if err != nil {
		return fmt.Errorf("open config file error, %s", err.Error())
	}

	fd, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("read config file error, %s", err.Error())
	}

	cfg = new(config.Config)

	cfg.App = new(config.AppConf)
	cfg.Command = new(config.CommandConf)

	if err := yaml.Unmarshal(fd, cfg); err != nil {
		return fmt.Errorf("load config file error, %s", err.Error())
	}

	return nil
}
