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

package util

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func RandomString(size int, chars string) string {
	buf := make([]byte, size)
	for idx := range buf {
		buf[idx] = chars[rand.Intn(len(chars))]
	}
	return string(buf)
}

const AlphaNum = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const Alpha = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const AlphaLower = "abcdefghijklmnopqrstuvwxyz"
