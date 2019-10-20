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

package execute

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/eclipse-iofog/iofog-go-sdk/pkg/apps"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"gopkg.in/yaml.v2"
)

func execDocument(header config.Header, namespace string, kindHandlers map[apps.Kind]func(string, string, []byte) (Executor, error)) (exe Executor, err error) {
	// Check namespace exists
	if len(header.Metadata.Namespace) > 0 {
		namespace = header.Metadata.Namespace
	}
	if _, err := config.GetNamespace(namespace); err != nil {
		return exe, err
	}

	subYamlBytes, err := yaml.Marshal(header.Spec)
	if err != nil {
		return exe, err
	}

	createExecutorf, found := kindHandlers[header.Kind]
	if !found {
		return exe, util.NewInputError("Invalid kind")
	}

	return createExecutorf(namespace, header.Metadata.Name, subYamlBytes)
}

func GetExecutorsFromYAML(inputFile, namespace string, kindHandlers map[apps.Kind]func(string, string, []byte) (Executor, error)) (executorsMap map[apps.Kind][]Executor, err error) {
	yamlFile, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return
	}

	r := bytes.NewReader(yamlFile)
	dec := yaml.NewDecoder(r)

	var raw yaml.MapSlice
	header := config.Header{
		Spec: raw,
	}

	// Generate all executors
	executorsMap = make(map[apps.Kind][]Executor)
	decodeErr := dec.Decode(&header)
	for decodeErr == nil {
		exe, err := execDocument(header, namespace, kindHandlers)
		if err != nil {
			return nil, err
		}
		executorsMap[header.Kind] = append(executorsMap[header.Kind], exe)
		decodeErr = dec.Decode(&header)
	}
	if decodeErr != io.EOF && decodeErr != nil {
		return nil, decodeErr
	}

	return
}
