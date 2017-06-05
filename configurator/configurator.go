package configurator

import (
	"github.com/pkg/errors"
	"os"
	"fmt"
)

// Configurator is struct of configurator
type Configurator struct {
	configPath string
	reader     *reader
	writer     *writer
}

// Load is load config
func (c *Configurator) Load(data interface{}) (err error) {
	return c.reader.read(c.configPath, data)
}

// Save is save config
func (c *Configurator) Save(data interface{}) (err error) {
	return c.writer.write(c.configPath, data)
}

// New is create Configurator
func New(configPath string) (configurator *Configurator, err error) {
	_, err = os.Stat(configPath)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("not exists config file (%v)", configPath))
	}
	return &Configurator{
		reader : newReader(),
		writer : newWriter(),
		configPath : configPath,
	}, nil
}
