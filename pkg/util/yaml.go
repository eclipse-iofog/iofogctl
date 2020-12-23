/*
 *  *******************************************************************************
 *  * Copyright (c) 2020 Edgeworx, Inc.
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package util

import (
	"io"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

func UnmarshalYAML(filename string, object interface{}) error {
	yamlFile, err := ioutil.ReadFile(filename)
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
	f, err := os.Create(filename)
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
