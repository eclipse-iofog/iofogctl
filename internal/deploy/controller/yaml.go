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

package deploycontroller

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"gopkg.in/yaml.v2"
)

func UnmarshallYAML(file []byte) (ctrl config.Controller, err error) {
	// Unmarshall the input file
	if err = yaml.UnmarshalStrict(file, &ctrl); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}

	// Fix replica count
	if ctrl.Replicas == 0 {
		ctrl.Replicas = 1
	}
	// Fix SSH port
	if ctrl.SSH.Port == 0 {
		ctrl.SSH.Port = 22
	}
	// Format file paths
	if ctrl.SSH.KeyFile, err = util.FormatPath(ctrl.SSH.KeyFile); err != nil {
		return
	}
	if ctrl.KubeConfig, err = util.FormatPath(ctrl.KubeConfig); err != nil {
		return
	}

	return
}

func Validate(ctrl config.Controller) error {
	if ctrl.Name == "" {
		return util.NewInputError("You must specify a non-empty value for name value of Controllers")
	}
	if ctrl.KubeConfig == "" && ((ctrl.SSH.Host != "localhost" && ctrl.SSH.Host != "127.0.0.1") && (ctrl.SSH.Host == "" || ctrl.SSH.User == "" || ctrl.SSH.KeyFile == "")) {
		return util.NewInputError("For Controllers you must specify non-empty values for EITHER kubeconfig OR host, user, and keyfile")
	}
	return nil
}
