package configurator

import (
        "github.com/pkg/errors"
	"github.com/BurntSushi/toml"
	"github.com/potix/pdns-record-updater/contexter"
	"encoding/json"
	"gopkg.in/yaml.v2"
	"path/filepath"
	"io/ioutil"
	"fmt"
)

type reader struct {
}

func (r *reader) read(configPath string) (*contexter.Context, error) {
	newContext := new(contexter.Context)
        ext := filepath.Ext(configPath)
        switch ext {
        case ".tml":
                fallthrough
        case ".toml":
                _, err := toml.DecodeFile(configPath, newContext)
                if err != nil {
                        return nil, errors.Wrap(err, fmt.Sprintf("can not decode file with toml (%v)", configPath))
                }
        case ".yml":
                fallthrough
        case ".yaml":
                buf, err := ioutil.ReadFile(configPath)
                if err != nil {
                        return nil, errors.Wrap(err, fmt.Sprintf("can not read file with yaml (%v)", configPath))
                }
                err = yaml.Unmarshal(buf, newContext)
                if err != nil {
                        return nil, errors.Wrap(err, fmt.Sprintf("can not decode file with yaml (%v)", configPath))
                }
        case ".jsn":
                fallthrough
        case ".json":
                buf, err := ioutil.ReadFile(configPath)
                if err != nil {
                        return nil, errors.Wrap(err, fmt.Sprintf("can not read file with json (%v)", configPath))
                }
                err = json.Unmarshal(buf, newContext)
                if err != nil {
                        return nil, errors.Wrap(err, fmt.Sprintf("can not decode file with json (%v)", configPath))
                }
        default:
                return nil, errors.Errorf("unexpected file extension (%v)", ext)
        }
	return newContext, nil
}

func newReader() (*reader) {
	return &reader{}
}
