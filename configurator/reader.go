package configurator

import (
        "github.com/pkg/errors"
	"github.com/BurntSushi/toml"
	"fmt"
)

type reader struct {
	config *Config
}

func (r *reader) read(configPath string) (*Config, error) {
	// XXX TODO support yaml json
	_, err := toml.DecodeFile(configPath, r.config)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("can not decode file with toml (%v)", configPath))
	}
	return r.config, nil
}

func newReader() (*reader) {
	return &reader{}
}
