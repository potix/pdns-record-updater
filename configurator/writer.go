package configurator

import (
        "github.com/pkg/errors"
	"github.com/BurntSushi/toml"
	"encoding/json"
	"gopkg.in/yaml.v2"
	"path/filepath"
	"io/ioutil"
	"bytes"
	"fmt"
)

type writer struct {
}

func (r *writer) write(configPath string, data interface{}) (error) {
        ext := filepath.Ext(configPath)
        switch ext {
        case ".tml":
                fallthrough
        case ".toml":
		var buffer bytes.Buffer
		encoder := toml.NewEncoder(&buffer)
		err := encoder.Encode(data)
		if err != nil {
                        return errors.Wrap(err, fmt.Sprintf("can not encode with toml (%v)", configPath))
		}
		err = ioutil.WriteFile(configPath, buffer.Bytes(), 0644)
		if err != nil {
                        return errors.Wrap(err, fmt.Sprintf("can not write file with toml (%v)", configPath))
		}
        case ".yml":
                fallthrough
        case ".yaml":
		y, err := yaml.Marshal(data)
                if err != nil {
                        return errors.Wrap(err, fmt.Sprintf("can not encode with yaml (%v)", configPath))
                }
		err = ioutil.WriteFile(configPath, y, 0644)
		if err != nil {
                        return errors.Wrap(err, fmt.Sprintf("can not write file with yaml (%v)", configPath))
		}
        case ".jsn":
                fallthrough
        case ".json":
		j, err := json.Marshal(data)
                if err != nil {
                        return errors.Wrap(err, fmt.Sprintf("can not encode with json (%v)", configPath))
                }
		err = ioutil.WriteFile(configPath, j, 0644)
		if err != nil {
                        return errors.Wrap(err, fmt.Sprintf("can not write file with json (%v)", configPath))
		}
        default:
                return errors.Errorf("unexpected file extension (%v)", ext)
        }
	return nil
}

func newWriter() (*writer) {
	return &writer{}
}
