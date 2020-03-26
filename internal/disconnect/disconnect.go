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

package disconnect

import (
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
)

type Options struct {
	Namespace string
}

func Execute(opt *Options) error {
	if err := config.DeleteNamespace(opt.Namespace); err != nil {
		return err
	}
	return config.Flush()
}
