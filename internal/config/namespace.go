package config

import (
	"errors"
	"os"
	"sort"

	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
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
	files, err := os.ReadDir(namespaceDirectory)
	util.Check(err)

	sort.Slice(files, func(i, j int) bool {
		infoI, err := files[i].Info()
		if err != nil {
			return false
		}
		infoJ, err := files[j].Info()
		if err != nil {
			return false
		}
		return infoI.ModTime().After(infoJ.ModTime())
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
		// Namespace has not been loaded from file, do so now
		namespaceHeader := iofogctlNamespace{}
		if err := util.UnmarshalYAML(getNamespaceFile(name), &namespaceHeader); err != nil {
			if os.IsNotExist(err) {
				return nil, util.NewNotFoundError(name)
			}
			return nil, err
		}
		ns, err := getNamespaceFromHeader(&namespaceHeader)
		if err != nil {
			return nil, err
		}
		namespaces[name] = ns
		return ns, flushNamespaces()
	}
	// Return Namespace from memory
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
	err = os.WriteFile(getNamespaceFile(name), marshal, 0644)
	if err != nil {
		return err
	}
	namespaces[name] = &newNamespace
	return nil
}

// UpdateUser modifies the user data of an existing namespace
func UpdateUser(name, accessToken, refreshToken string) error {
	// Fetch the existing namespace
	ns, err := getNamespace(name)
	if err != nil {
		return err // Namespace not found
	}

	if ns.KubernetesControlPlane != nil {
		ns.KubernetesControlPlane.IofogUser.AccessToken = accessToken
		ns.KubernetesControlPlane.IofogUser.RefreshToken = refreshToken
	}
	if ns.RemoteControlPlane != nil {
		ns.RemoteControlPlane.IofogUser.AccessToken = accessToken
		ns.RemoteControlPlane.IofogUser.RefreshToken = refreshToken
	}
	if ns.LocalControlPlane != nil {
		ns.LocalControlPlane.IofogUser.AccessToken = accessToken
		ns.LocalControlPlane.IofogUser.RefreshToken = refreshToken
	}

	// Marshal the updated namespace into YAML
	marshal, err := getNamespaceYAMLFile(ns)
	if err != nil {
		return err // Error in marshaling
	}

	// Write the updated YAML data back to the file
	err = os.WriteFile(getNamespaceFile(name), marshal, 0644)
	if err != nil {
		return err // Error in writing to the file
	}

	// Update the in-memory cache
	namespaces[name] = ns

	return nil
}

// DeleteNamespace removes a namespace including all the resources within it
func DeleteNamespace(name string) error {
	// Reset default namespace if required
	if name == conf.DefaultNamespace {
		if err := SetDefaultNamespace("default"); err != nil {
			msg := "failed to reconfigure default namespace"
			return errors.New(msg)
		}
	}

	filename := getNamespaceFile(name)
	if err := os.Remove(filename); err != nil {
		msg := "could not delete namespace file " + filename
		return util.NewNotFoundError(msg)
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
	if err := os.Rename(getNamespaceFile(name), getNamespaceFile(newName)); err != nil {
		return err
	}
	if name == conf.DefaultNamespace {
		return SetDefaultNamespace(newName)
	}

	return nil
}
