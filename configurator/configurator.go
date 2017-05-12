package configurator

import (
	"os"
)

// Configurator is struct of configurator
type Configurator struct {
	configPath string
	reader     *reader
}

// Load is load config
func (c *Configurator) Load() (config *Config, err error) {
	return Configurator.reader.read()
}

// New is create Configurator
func New(onfigPath string) (configurator *Configurator, err error) {
	configurator = new(Configurator)
	_, err := os.Stat(configPAth)
	if err != nil {
		return nil, err
	}
	return &Configurator{
		reader : reader.New(),
		configPath : configPath,
	}, nil
}
