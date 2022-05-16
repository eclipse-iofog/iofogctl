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

package install

import (
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

// Microservice names
const (
	controller = "controller"
)

type microservice struct {
	name       string
	IP         string
	ports      []int32
	replicas   int32
	containers []container
	rbacRules  []rbacv1.PolicyRule
}

type container struct {
	name            string
	image           string
	imagePullPolicy string
	args            []string
	readinessProbe  *corev1.Probe
	env             []corev1.EnvVar
	command         []string
	ports           []corev1.ContainerPort
}

func newOperatorMicroservice() *microservice {
	return &microservice{
		name:     "iofog-operator",
		ports:    []int32{60000},
		replicas: 1,
		rbacRules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{
					"rbac.authorization.k8s.io",
				},
				Resources: []string{
					"roles",
					"rolebindings",
				},
				Verbs: []string{
					"*",
				},
			},
			{
				APIGroups: []string{
					"iofog.org",
				},
				Resources: []string{
					"apps",
					"applications",
					"applications/status",
					"controlplanes",
					"apps/status",
					"controlplanes/status",
				},
				Verbs: []string{
					"list",
					"get",
					"watch",
					"update",
				},
			},
			{
				APIGroups: []string{
					"apps",
				},
				Resources: []string{
					"deployments",
				},
				Verbs: []string{
					"*",
				},
			},
			{
				APIGroups: []string{
					"coordination.k8s.io",
				},
				Resources: []string{
					"leases",
				},
				Verbs: []string{
					"*",
				},
			},
			{
				APIGroups: []string{
					"",
				},
				Resources: []string{
					"pods",
					"configmaps",
					"configmaps/status",
					"events",
					"serviceaccounts",
					"services",
					"persistentvolumeclaims",
					"secrets",
				},
				Verbs: []string{
					"*",
				},
			},
		},
		containers: []container{
			{
				name:            "iofog-operator",
				image:           util.GetOperatorImage(),
				imagePullPolicy: "Always",
				env: []corev1.EnvVar{
					{
						Name: "WATCH_NAMESPACE",
						ValueFrom: &corev1.EnvVarSource{
							FieldRef: &corev1.ObjectFieldSelector{
								FieldPath: "metadata.namespace",
							},
						},
					},
					{
						Name: "POD_NAME",
						ValueFrom: &corev1.EnvVarSource{
							FieldRef: &corev1.ObjectFieldSelector{
								FieldPath: "metadata.name",
							},
						},
					},
					{
						Name:  "OPERATOR_NAME",
						Value: "iofog-operator",
					},
				},
				ports: []corev1.ContainerPort{
					{
						ContainerPort: int32(60000),
						Name:          "metrics",
					},
				},
				command: []string{
					"iofog-operator",
				},
				args: []string{
					"--enable-leader-election",
				},
			},
		},
	}
}
