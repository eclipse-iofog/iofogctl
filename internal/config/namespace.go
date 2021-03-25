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

	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

func SetDefaultNamespace(name string) (err error) {
	if name == pkg.conf.DefaultNamespace {
		return
	}
	// Check exists
	for _, n := range GetNamespaces() {
		if n == name {
			pkg.conf.DefaultNamespace = name
			return flushShared()
		}
	}
	return util.NewNotFoundError(name)
}

// GetNamespaces returns all namespaces in config
func GetNamespaces() (namespaces []string) {
	files := []os.FileInfo{}
	for _, dir := range pkg.nsDirs {
		nsFiles, err := ioutil.ReadDir(dir)
		util.Check(err)
		files = append(files, nsFiles...)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Before(files[j].ModTime())
	})

	for _, file := range files {
		name := util.Before(file.Name(), ".yaml")
		if name != pkg.detachedNamespace {
			namespaces = append(namespaces, name)
		}
	}
	return
}

func GetDefaultNamespaceName() string {
	return pkg.conf.DefaultNamespace
}

func findNamespaceVersion(name string) (version string, err error) {
	// TODO: Replace this with LS dir for ns file
	for _, vers := range pkg.supportedVersions {
		if _, err = getNamespace(name, vers); err == nil {
			version = vers
			break
		}
	}
	return version, err
}

func getNamespace(name, version string) (*rsc.Namespace, error) {
	namespace, ok := pkg.nsIndex[version][name]
	if !ok {
		// Namespace has not been loaded from file, do so now
		namespaceHeader := iofogctlNamespace{}
		if err := util.UnmarshalYAML(getNamespaceFile(name, version), &namespaceHeader); err != nil {
			if os.IsNotExist(err) {
				return nil, util.NewNotFoundError(name)
			}
			return nil, err
		}
		ns, err := getNamespaceFromHeader(&namespaceHeader)
		if err != nil {
			return nil, err
		}
		pkg.nsIndex[version][name] = ns
		return ns, flushNamespaces()
	}
	// Return Namespace from memory
	return namespace, nil
}

// GetNamespace returns the namespace
func GetNamespace(name string) (ns *rsc.Namespace, err error) {
	version, err := findNamespaceVersion(name)
	if err != nil {
		return nil, err
	}
	return getNamespace(name, version)
}

// AddNamespace adds a new namespace to the config
func AddNamespace(name, created string) error {
	// Check collision
	for _, n := range GetNamespaces() {
		if n == name {
			return util.NewConflictError(name)
		}
	}
	return addNamespace(name, created, pkg.latestVersion)
}

func addNamespace(name, created, version string) error {
	newNamespace := rsc.Namespace{
		Name:    name,
		Created: created,
	}

	// Write namespace file
	// Marshal the runtime data
	marshal, err := getNamespaceYAMLFile(&newNamespace, version)
	if err != nil {
		return err
	}
	// Overwrite the file
	err = ioutil.WriteFile(getNamespaceFile(name, version), marshal, 0644)
	if err != nil {
		return err
	}
	pkg.nsIndex[version][name] = &newNamespace
	return nil
}

// DeleteNamespace removes a namespace including all the resources within it
func DeleteNamespace(name string) error {
	// Reset default namespace if required
	if name == pkg.conf.DefaultNamespace {
		if err := SetDefaultNamespace("default"); err != nil {
			msg := "failed to reconfigure default namespace"
			return errors.New(msg)
		}
	}
	// Find ns version
	version, err := findNamespaceVersion(name)
	if err != nil {
		return err
	}

	return deleteNamespace(name, version)
}

func deleteNamespace(name, version string) error {
	filename := getNamespaceFile(name, version)
	if err := os.Remove(filename); err != nil {
		msg := "could not delete namespace file " + filename
		return util.NewNotFoundError(msg)
	}

	delete(pkg.nsIndex[version], name)

	return nil
}

// RenameNamespace renames a namespace
func RenameNamespace(name, newName string) error {
	ns, err := getNamespace(name, pkg.latestVersion)
	if err != nil {
		util.PrintError("Could not find namespace " + name)
		return err
	}
	// Find ns version
	version, err := findNamespaceVersion(name)
	if err != nil {
		return err
	}
	return renameNamespace(ns, newName, version)
}

func renameNamespace(ns *rsc.Namespace, newName, version string) error {
	name := ns.Name
	ns.Name = newName
	if err := os.Rename(getNamespaceFile(name, version), getNamespaceFile(newName, version)); err != nil {
		return err
	}
	if name == pkg.conf.DefaultNamespace {
		return SetDefaultNamespace(newName)
	}

	return nil
}
