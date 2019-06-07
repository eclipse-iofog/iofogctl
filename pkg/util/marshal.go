package util

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

func UnmarshalYAML(filename string, object interface{}) error {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFile, object)
	if err != nil {
		return err
	}

	return nil
}
