package configurator

import (
	"github.com/BurntSushi/toml"
        "github.com/pkg/errors"
)

type reader struct {
	config *Config
}

func (r *reader) read(configPath *string) (*Config, error) {
	var err error
	_, err = toml.DecodeFile(configPath, reader.config)
	if err != nil {
		return nil, err
	}
	return reader.config, nil
}

func new() {
	return &reader{}
}
