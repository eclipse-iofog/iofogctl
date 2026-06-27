package util

import (
	"io"
	"os"

	"gopkg.in/yaml.v2"
)

func UnmarshalYAML(filename string, object interface{}) error {
	yamlFile, err := ReadUserFile(filename)
	if err != nil {
		return err
	}
	err = yaml.UnmarshalStrict(yamlFile, object)
	if err != nil {
		return err
	}

	return nil
}

func printYAML(writer io.Writer, obj interface{}) error {
	marshal, err := yaml.Marshal(&obj)
	if err != nil {
		return err
	}
	_, err = writer.Write(marshal)
	if err != nil {
		return err
	}
	return nil
}

func FPrint(obj interface{}, filename string) error {
	f, err := CreateUserFile(filename, FilePerm)
	defer Log(f.Close)
	if err != nil {
		return err
	}
	err = printYAML(f, obj)
	if err != nil {
		return err
	}
	return nil
}

func Print(obj interface{}) error {
	err := printYAML(os.Stdout, obj)
	if err != nil {
		return err
	}
	return nil
}
