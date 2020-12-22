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

package config

import (
	"errors"
	"io/ioutil"
	"os"
	"sort"

	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

func SetDefaultNamespace(name string) (err error) {
	if name == conf.DefaultNamespace {
		return
	}
	// Check exists
	for _, n := range GetNamespaces() {
		if n == name {
			conf.DefaultNamespace = name
			return flushShared()
		}
	}
	return util.NewNotFoundError(name)
}

// GetNamespaces returns all namespaces in config
func GetNamespaces() (namespaces []string) {
	files, err := ioutil.ReadDir(namespaceDirectory)
	util.Check(err)

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Before(files[j].ModTime())
	})

	for _, file := range files {
		name := util.Before(file.Name(), ".yaml")
		if name != detachedNamespace {
			namespaces = append(namespaces, name)
		}
	}
	return
}

func GetDefaultNamespaceName() string {
	return conf.DefaultNamespace
}

func getNamespace(name string) (*rsc.Namespace, error) {
	namespace, ok := namespaces[name]
	if !ok {
		namespaceHeader := iofogctlNamespace{}
		if err := util.UnmarshalYAML(getNamespaceFile(name), &namespaceHeader); err != nil {
			if os.IsNotExist(err) {
				return nil, util.NewNotFoundError(name)
			}
			return nil, err
		}
		ns, err := getNamespaceFromHeader(namespaceHeader)
		if err != nil {
			return nil, err
		}
		namespaces[name] = &ns
		return &ns, flushNamespaces()
	}
	return namespace, nil
}

// GetNamespace returns the namespace
func GetNamespace(namespace string) (*rsc.Namespace, error) {
	ns, err := getNamespace(namespace)
	if err != nil {
		return nil, err
	}
	return ns, nil
}

// AddNamespace adds a new namespace to the config
func AddNamespace(name, created string) error {
	// Check collision
	for _, n := range GetNamespaces() {
		if n == name {
			return util.NewConflictError(name)
		}
	}

	newNamespace := rsc.Namespace{
		Name:    name,
		Created: created,
	}

	// Write namespace file
	// Marshal the runtime data
	marshal, err := getNamespaceYAMLFile(&newNamespace)
	if err != nil {
		return err
	}
	// Overwrite the file
	err = ioutil.WriteFile(getNamespaceFile(name), marshal, 0644)
	if err != nil {
		return err
	}
	namespaces[name] = &newNamespace
	return nil
}

// DeleteNamespace removes a namespace including all the resources within it
func DeleteNamespace(name string) error {
	// Reset default namespace if required
	if name == conf.DefaultNamespace {
		if err := SetDefaultNamespace("default"); err != nil {
			return errors.New("Failed to reconfigure default namespace")
		}
	}

	filename := getNamespaceFile(name)
	if err := os.Remove(filename); err != nil {
		return util.NewNotFoundError("Could not delete namespace file " + filename)
	}

	delete(namespaces, name)

	return nil
}

// RenameNamespace renames a namespace
func RenameNamespace(name, newName string) error {
	ns, err := getNamespace(name)
	if err != nil {
		util.PrintError("Could not find namespace " + name)
		return err
	}
	ns.Name = newName
	if err = os.Rename(getNamespaceFile(name), getNamespaceFile(newName)); err != nil {
		return err
	}
	if name == conf.DefaultNamespace {
		return SetDefaultNamespace(newName)
	}

	return nil
}
