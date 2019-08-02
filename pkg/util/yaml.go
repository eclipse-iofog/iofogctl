/*
 *  *******************************************************************************
 *  * Copyright (c) 2019 Edgeworx, Inc.
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
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
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

func Print(obj interface{}, filename string) error {
	marshal, err := yaml.Marshal(&obj)
	if err != nil {
		return err
	}
	if filename != "" {
		if err = ioutil.WriteFile(filename, marshal, 0644); err != nil {
			return err
		}
	} else {
		_, err = os.Stdout.Write(marshal)
		if err != nil {
			return err
		}
	}
	return nil
}
