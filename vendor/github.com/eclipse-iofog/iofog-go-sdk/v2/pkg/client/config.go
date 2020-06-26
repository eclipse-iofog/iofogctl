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

package client

type Protocol = string

const (
	TCP = "tcp"
	HTTP = "http"
)

func (clt *Client) PutPublicPortHost(protocol Protocol, host string) (err error) {
	_, err = clt.doRequest("PUT", "/config", newPublicPortHostRequest(protocol, host))
	return
}

func (clt *Client) PutDefaultProxy(address string) (err error) {
	_, err = clt.doRequest("PUT", "/config", newDefaultProxyRequest(address))
	return
}
