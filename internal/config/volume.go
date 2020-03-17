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

package config

import (
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

func AddVolume(namespace string, volume Volume) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	if _, err := GetVolume(namespace, volume.Name); err == nil {
		return util.NewConflictError(namespace + "/" + volume.Name)
	}

	mux.Lock()
	ns.Volumes = append(ns.Volumes, volume)
	mux.Unlock()
	return nil
}

func DeleteVolume(namespace, name string) error {
	ns, err := getNamespace(namespace)
	if err != nil {
		return err
	}
	for idx := range ns.Volumes {
		if ns.Volumes[idx].Name == name {
			mux.Lock()
			ns.Volumes = append(ns.Volumes[:idx], ns.Volumes[idx+1:]...)
			mux.Unlock()
			return nil
		}
	}
	return util.NewNotFoundError(ns.Name + "/" + name)
}

func GetVolumes(namespace string) ([]Volume, error) {
	ns, err := getNamespace(namespace)
	if err != nil {
		return nil, err
	}
	return ns.Volumes, nil
}

func GetVolume(namespace, name string) (agent Volume, err error) {
	ns, err := getNamespace(namespace)
	if err != nil {
		return
	}
	for _, ag := range ns.Volumes {
		if ag.Name == name {
			agent = ag
			return
		}
	}

	err = util.NewNotFoundError(namespace + "/" + name)
	return
}
