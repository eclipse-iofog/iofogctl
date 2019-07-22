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

package install

import (
	"github.com/GeertJohan/go.rice"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

var installAgentScript string
var waitAgentScript string
var installControllerScript string

func init() {

	assets, err := rice.FindBox("../../../assets")
	util.Check(err)

	installAgentScript, err = assets.String("install_agent.sh")
	util.Check(err)

	waitAgentScript, err = assets.String("wait_agent.sh")
	util.Check(err)

	installControllerScript, err = assets.String("install_controller.sh")
	util.Check(err)
}
