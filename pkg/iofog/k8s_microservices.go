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

package iofog

import (
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type microservice struct {
	name       string
	port       int
	replicas   int32
	containers []container
}

type container struct {
	name            string
	image           string
	imagePullPolicy string
	args            []string
	readinessProbe  *v1.Probe
	env             []v1.EnvVar
	command         []string
	ports           []v1.ContainerPort
}

var controllerMicroservice = microservice{
	name:     "controller",
	port:     51121,
	replicas: 1,
	containers: []container{
		{
			name:            "controller",
			image:           "edgeworx/controller-k8s:latest",
			imagePullPolicy: "Always",
			readinessProbe: &v1.Probe{
				Handler: v1.Handler{
					HTTPGet: &v1.HTTPGetAction{
						Path: "/api/v3/status",
						Port: intstr.FromInt(51121),
					},
				},
				InitialDelaySeconds: 1,
				PeriodSeconds:       4,
				FailureThreshold:    3,
			},
		},
	},
}

var connectorMicroservice = microservice{
	name:     "connector",
	port:     8080,
	replicas: 1,
	containers: []container{
		{
			name:            "connector",
			image:           "iofog/connector:dev",
			imagePullPolicy: "Always",
		},
	},
}

var schedulerMicroservice = microservice{
	name:     "scheduler",
	replicas: 1,
	containers: []container{
		{
			name:            "scheduler",
			image:           "iofog/iofog-scheduler:develop",
			imagePullPolicy: "Always",
		},
	},
}

var operatorMicroservice = microservice{
	name:     "operator",
	port:     60000,
	replicas: 1,
	containers: []container{
		{
			name:            "operator",
			image:           "iofog/iofog-operator:develop",
			imagePullPolicy: "Always",
			readinessProbe: &v1.Probe{
				Handler: v1.Handler{
					Exec: &v1.ExecAction{
						Command: []string{
							"stat",
							"/tmp/operator-sdk-ready",
						},
					},
				},
				InitialDelaySeconds: 4,
				PeriodSeconds:       10,
				FailureThreshold:    1,
			},
			env: []v1.EnvVar{
				{
					Name: "WATCH_NAMESPACE",
					ValueFrom: &v1.EnvVarSource{
						FieldRef: &v1.ObjectFieldSelector{
							FieldPath: "metadata.namespace",
						},
					},
				},
				{
					Name: "POD_NAME",
					ValueFrom: &v1.EnvVarSource{
						FieldRef: &v1.ObjectFieldSelector{
							FieldPath: "metadata.name",
						},
					},
				},
				{
					Name:  "OPERATOR_NAME",
					Value: "operator",
				},
			},
			ports: []v1.ContainerPort{
				{
					ContainerPort: int32(60000),
					Name:          "metrics",
				},
			},
			command: []string{
				"iofog-operator",
			},
		},
	},
}

var kubeletMicroservice = microservice{
	name:     "kubelet",
	port:     60000,
	replicas: 1,
	containers: []container{
		{
			name:            "kubelet",
			image:           "iofog/iofog-kubelet:develop",
			imagePullPolicy: "Always",
			args: []string{
				"--namespace",
				"",
				"--iofog-token",
				"",
				"--iofog-url",
				"",
			},
		},
	},
}
