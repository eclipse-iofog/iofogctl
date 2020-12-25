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

package resource

import (
	"encoding/base64"
)

// IofogUser contains information about users registered against a controller
type IofogUser struct {
	Name     string `yaml:"name,omitempty"`
	Surname  string `yaml:"surname,omitempty"`
	Email    string `yaml:"email,omitempty"`
	Password string `yaml:"password,omitempty"`
}

func (user *IofogUser) EncodePassword() {
	user.Password = encodeBase64(user.Password)
}

func (user IofogUser) GetRawPassword() string {
	buf, err := base64.StdEncoding.DecodeString(user.Password)
	if err != nil {
		return user.Password
	}
	return string(buf)
}

func encodeBase64(raw string) string {
	return base64.StdEncoding.EncodeToString([]byte(raw))
}
