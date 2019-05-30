package iofog

import (
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func newService(namespace string, ms microservice) *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ms.name,
			Namespace: namespace,
			Labels: map[string]string{
				"name": ms.name,
			},
		},
		Spec: v1.ServiceSpec{
			Type: "LoadBalancer",
			Ports: []v1.ServicePort{
				{
					Name:       "controller",
					Port:       int32(ms.port),
					TargetPort: intstr.FromInt(ms.port),
				},
			},
		},
	}
}

func newDeployment(namespace string, ms microservice) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ms.name,
			Namespace: namespace,
			Labels: map[string]string{
				"name": ms.name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &ms.replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": ms.name,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"name": ms.name,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            ms.name,
							Image:           ms.image,
							ImagePullPolicy: v1.PullPolicy(ms.imagePullPolicy),
						},
					},
				},
			},
		},
	}
}
