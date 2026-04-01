/*
 *  *******************************************************************************
 *  * Copyright (c) 2023 Contributors to the Eclipse ioFog Project
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package exec

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Options struct {
	Resource  string
	Name      string
	Namespace string
}

func NewExecutor(opt *Options) (execute.Executor, error) {
	switch opt.Resource {
	case "microservice":
		return newMicroserviceExecutor(opt.Namespace, opt.Name), nil
	case "agent":
		return newAgentExecutor(opt.Namespace, opt.Name), nil
	default:
		return nil, util.NewInputError(fmt.Sprintf("Unknown resources: %s", opt.Resource))
	}
}
