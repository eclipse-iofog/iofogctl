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

package install

import (
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Microservice names
const (
	kubelet    = "kubelet"
	operator   = "operator"
	controller = "controller"
	connector  = "connector"
)

type microservice struct {
	name        string
	IP          string
	ports       []int32
	serviceType string
	replicas    int32
	containers  []container
	rbacRules   []rbacv1.PolicyRule
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

func newControllerMicroservice() *microservice {
	return &microservice{
		name:        "controller",
		ports:       []int32{iofog.ControllerPort, 80},
		replicas:    1,
		serviceType: string(corev1.ServiceTypeLoadBalancer),
		containers: []container{
			{
				name:            "controller",
				image:           "iofog/controller:" + util.GetControllerTag(),
				imagePullPolicy: "Always",
				readinessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						HTTPGet: &corev1.HTTPGetAction{
							Path: "/api/v3/status",
							Port: intstr.FromInt(iofog.ControllerPort),
						},
					},
					InitialDelaySeconds: 1,
					PeriodSeconds:       4,
					FailureThreshold:    3,
				},
			},
		},
	}
}

func newConnectorMicroservice() *microservice {
	return &microservice{
		name: "connector",
		ports: []int32{
			iofog.ConnectorPort,
			6000, 6001, 6002, 6003, 6004, 6005, 6006, 6007, 6008, 6009,
			6010, 6011, 6012, 6013, 6014, 6015, 6016, 6017, 6018, 6019,
			6020, 6021, 6022, 6023, 6024, 6025, 6026, 6027, 6028, 6029,
			6030, 6031, 6032, 6033, 6034, 6035, 6036, 6037, 6038, 6039,
			6040, 6041, 6042, 6043, 6044, 6045, 6046, 6047, 6048, 6049,
			6050,
		},
		replicas:    1,
		serviceType: string(corev1.ServiceTypeLoadBalancer),
		containers: []container{
			{
				name:            "connector",
				image:           "iofog/connector:" + util.GetConnectorTag(),
				imagePullPolicy: "Always",
			},
		},
	}
}

func newSchedulerMicroservice() *microservice {
	return &microservice{
		name:     "scheduler",
		replicas: 1,
		containers: []container{
			{
				name:            "scheduler",
				image:           "iofog/iofog-scheduler:" + util.GetSchedulerTag(),
				imagePullPolicy: "Always",
			},
		},
	}
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
		},
		containers: []container{
			{
				name:            "iofog-operator",
				image:           "iofog/iofog-operator:" + util.GetOperatorTag(),
				imagePullPolicy: "Always",
				readinessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						Exec: &corev1.ExecAction{
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
			},
		},
	}
}

func newKubeletMicroservice() *microservice {
	return &microservice{
		name:     "kubelet",
		ports:    []int32{60000},
		replicas: 1,
		containers: []container{
			{
				name:            "kubelet",
				image:           "iofog/iofog-kubelet:" + util.GetKubeletTag(),
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
}
