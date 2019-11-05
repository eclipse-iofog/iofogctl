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

package connectcontroller

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"gopkg.in/yaml.v2"
)

func unmarshallYAML(file []byte) (ctrl config.Controller, err error) {
	// Unmarshall the input file
	if err = yaml.UnmarshalStrict(file, &ctrl); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}

	// Fix replica count
	if ctrl.Kube.Replicas == 0 {
		ctrl.Kube.Replicas = 1
	}
	// Fix SSH port
	if ctrl.SSH.Port == 0 {
		ctrl.SSH.Port = 22
	}
	// Format file paths
	if ctrl.SSH.KeyFile, err = util.FormatPath(ctrl.SSH.KeyFile); err != nil {
		return
	}
	if ctrl.Kube.Config, err = util.FormatPath(ctrl.Kube.Config); err != nil {
		return
	}

	return
}

func Validate(ctrl config.Controller) error {
	if ctrl.Name == "" {
		return util.NewInputError("You must specify a non-empty value for name value of Controllers")
	}
	if ctrl.Kube.Config == "" && ((ctrl.SSH.Host != "localhost" && ctrl.SSH.Host != "127.0.0.1") && ctrl.SSH.Host == "") {
		return util.NewInputError("To connect, for Controllers you must specify non-empty values for EITHER kubeconfig OR host")
	}
	return nil
}
