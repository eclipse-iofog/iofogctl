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

package describe

import (
	yaml "gopkg.in/yaml.v2"
	"os"
)

func print(obj interface{}) error {
	marshal, err := yaml.Marshal(&obj)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(marshal)
	if err != nil {
		return err
	}
	return nil
}
