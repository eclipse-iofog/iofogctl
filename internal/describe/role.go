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

package describe

import (
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type roleExecutor struct {
	namespace string
	name      string
	filename  string
}

func newRoleExecutor(namespace, name, filename string) *roleExecutor {
	return &roleExecutor{
		namespace: namespace,
		name:      name,
		filename:  filename,
	}
}

func (exe *roleExecutor) GetName() string {
	return exe.name
}

func (exe *roleExecutor) Execute() error {
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	role, err := clt.GetRole(exe.name)
	if err != nil {
		return err
	}

	spec := rsc.Role{
		Name:  role.Name,
		Kind:  role.Kind,
		Rules: convertRules(role.Rules),
	}

	header := config.Header{
		APIVersion: config.LatestAPIVersion,
		Kind:       config.RoleKind,
		Metadata: config.HeaderMetadata{
			Namespace: exe.namespace,
			Name:      exe.name,
		},
		Spec: spec,
	}

	if exe.filename == "" {
		if err := util.Print(header); err != nil {
			return err
		}
	} else {
		if err := util.FPrint(header, exe.filename); err != nil {
			return err
		}
	}
	return nil
}

func convertRules(rules []client.RBACRule) []rsc.RBACRule {
	out := make([]rsc.RBACRule, len(rules))
	for i, r := range rules {
		out[i] = rsc.RBACRule{
			APIGroups:     r.APIGroups,
			Resources:     r.Resources,
			Verbs:         r.Verbs,
			ResourceNames: r.ResourceNames,
		}
	}
	return out
}
