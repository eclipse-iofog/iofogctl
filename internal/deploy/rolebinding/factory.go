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

package deployrolebinding

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"gopkg.in/yaml.v2"
)

type Options struct {
	Namespace string
	Yaml      []byte
	Name      string
}

type executor struct {
	binding   rsc.RoleBinding
	namespace string
	name      string
}

func (exe *executor) GetName() string {
	return exe.name
}

func (exe *executor) Execute() error {
	util.SpinStart(fmt.Sprintf("Deploying rolebinding %s", exe.GetName()))
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	if _, err = clt.GetRoleBinding(exe.name); err != nil {
		return exe.createRoleBinding(clt)
	}
	return exe.updateRoleBinding(clt)
}

func (exe *executor) createRoleBinding(clt *client.Client) error {
	req := &client.RoleBindingCreateRequest{
		Name:     exe.name,
		Kind:     exe.binding.Kind,
		RoleRef:  toClientRoleRef(exe.binding.RoleRef),
		Subjects: toClientSubjects(exe.binding.Subjects),
	}
	_, err := clt.CreateRoleBinding(req)
	return err
}

func (exe *executor) updateRoleBinding(clt *client.Client) error {
	req := client.RoleBindingUpdateRequest{
		Name:     &exe.name,
		Kind:     exe.binding.Kind,
		RoleRef:  toClientRoleRef(exe.binding.RoleRef),
		Subjects: toClientSubjects(exe.binding.Subjects),
	}
	_, err := clt.UpdateRoleBinding(exe.name, &req)
	return err
}

func toClientRoleRef(ref rsc.RoleRef) client.RoleRef {
	return client.RoleRef{
		Kind:     ref.Kind,
		Name:     ref.Name,
		APIGroup: ref.APIGroup,
	}
}

func toClientSubjects(subjects []rsc.Subject) []client.Subject {
	out := make([]client.Subject, len(subjects))
	for i, s := range subjects {
		out[i] = client.Subject{
			Kind:     s.Kind,
			Name:     s.Name,
			APIGroup: s.APIGroup,
		}
	}
	return out
}

func NewExecutor(opt Options) (execute.Executor, error) {
	ns, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return nil, err
	}
	if len(ns.GetControllers()) == 0 {
		return nil, util.NewInputError("This namespace does not have a Controller. You must first deploy a Controller before deploying RoleBindings")
	}

	var binding rsc.RoleBinding
	if err = yaml.UnmarshalStrict(opt.Yaml, &binding); err != nil {
		return nil, util.NewUnmarshalError(err.Error())
	}

	name := opt.Name
	if name == "" {
		name = binding.Name
	}
	if name == "" {
		return nil, util.NewInputError("Name must be specified")
	}

	return &executor{
		namespace: opt.Namespace,
		binding:   binding,
		name:      name,
	}, nil
}
