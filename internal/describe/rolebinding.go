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

type roleBindingExecutor struct {
	namespace string
	name      string
	filename  string
}

func newRoleBindingExecutor(namespace, name, filename string) *roleBindingExecutor {
	return &roleBindingExecutor{
		namespace: namespace,
		name:      name,
		filename:  filename,
	}
}

func (exe *roleBindingExecutor) GetName() string {
	return exe.name
}

func (exe *roleBindingExecutor) Execute() error {
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	binding, err := clt.GetRoleBinding(exe.name)
	if err != nil {
		return err
	}

	spec := rsc.RoleBinding{
		Name:     binding.Name,
		Kind:     binding.Kind,
		RoleRef:  convertRoleRef(binding.RoleRef),
		Subjects: convertSubjects(binding.Subjects),
	}

	header := config.Header{
		APIVersion: config.LatestAPIVersion,
		Kind:       config.RoleBindingKind,
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

func convertRoleRef(ref client.RoleRef) rsc.RoleRef {
	return rsc.RoleRef{
		Kind:     ref.Kind,
		Name:     ref.Name,
		APIGroup: ref.APIGroup,
	}
}

func convertSubjects(subjects []client.Subject) []rsc.Subject {
	out := make([]rsc.Subject, len(subjects))
	for i, s := range subjects {
		out[i] = rsc.Subject{
			Kind:     s.Kind,
			Name:     s.Name,
			APIGroup: s.APIGroup,
		}
	}
	return out
}
