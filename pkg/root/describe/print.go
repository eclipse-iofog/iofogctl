package describe

import (
	yaml "gopkg.in/yaml.v2"
	"os"
)

func print(obj interface{}) error {
	marshal, err := yaml.Marshal(&obj)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(marshal)
	if err != nil {
		return err
	}
	return nil
}
