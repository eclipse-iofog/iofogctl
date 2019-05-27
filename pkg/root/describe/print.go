package describe

import (
	"os"
	yaml "gopkg.in/yaml.v2"
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