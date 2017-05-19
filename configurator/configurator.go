package configurator

import (
	"github.com/pkg/errors"
	"github.com/potix/pdns-record-updater/contexter"
	"os"
	"fmt"
)

// Configurator is struct of configurator
type Configurator struct {
	configPath string
	reader     *reader
}

// Load is load config
func (c *Configurator) Load() (context *contexter.Context, err error) {
	return c.reader.read(c.configPath)
}

// New is create Configurator
func New(configPath string) (configurator *Configurator, err error) {
	_, err = os.Stat(configPath)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("not exists config file (%v)", configPath))
	}
	return &Configurator{
		reader : newReader(),
		configPath : configPath,
	}, nil
}
