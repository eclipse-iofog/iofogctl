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

package deleteconnector

func (exe executor) k8sRemove() error {
	return nil
	//	// Find the requested controller
	//	cnct, err := config.GetConnector(exe.namespace, exe.name)
	//	if err != nil {
	//		return err
	//	}
	//
	//	// Instantiate Kubernetes object
	//	k8s, err := install.NewKubernetes(cnct.Kube.Config, exe.namespace)
	//	if err != nil {
	//		return err
	//	}
	//
	//	// Delete Connector on cluster
	//	err = k8s.DeleteConnector(exe.name)
	//	if err != nil {
	//		return err
	//	}
	//
	//	return nil
}
