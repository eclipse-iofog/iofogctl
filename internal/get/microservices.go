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

package get

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type microserviceExecutor struct {
	namespace string
}

func newMicroserviceExecutor(namespace string) *microserviceExecutor {
	a := &microserviceExecutor{}
	a.namespace = namespace
	return a
}

func (exe *microserviceExecutor) Execute() error {
	microservices, err := config.GetMicroservices(exe.namespace)
	if err != nil {
		return err
	}

	// Generate table and headers
	table := make([][]string, len(microservices)+1)
	headers := []string{"MICROSERVICE", "STATUS", "AGE"}
	table[0] = append(table[0], headers...)

	// Populate rows
	for idx, ms := range microservices {
		age, err := util.ElapsedUTC(ms.Created, util.NowUTC())
		if err != nil {
			return err
		}
		row := []string{
			ms.Name,
			"-",
			age,
		}
		table[idx+1] = append(table[idx+1], row...)
	}

	// Print the table
	err = print(table)
	if err != nil {
		return err
	}

	return nil
}
