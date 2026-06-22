package install

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func operatorDeploymentLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       "iofog",
		"app.kubernetes.io/instance":   "iofog",
		"app.kubernetes.io/component":  "iofog-operator",
		"app.kubernetes.io/managed-by": util.GetCliBinaryName(),
		"iofog.org/component":          "iofog-operator",
	}
}

func newDeployment(namespace string, ms *microservice) *appsv1.Deployment {
	maxUnavailable := intstr.FromInt(0)
	maxSurge := intstr.FromInt(1)
	strategy := appsv1.DeploymentStrategy{
		Type: appsv1.RollingUpdateDeploymentStrategyType,
		RollingUpdate: &appsv1.RollingUpdateDeployment{
			MaxUnavailable: &maxUnavailable,
			MaxSurge:       &maxSurge,
		},
	}
	depLabels := map[string]string{"name": ms.name}
	podLabels := map[string]string{"name": ms.name}
	if ms.name == "iofog-operator" {
		for k, v := range operatorDeploymentLabels() {
			depLabels[k] = v
			podLabels[k] = v
		}
	}
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ms.name,
			Namespace: namespace,
			Labels:    depLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &ms.replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": ms.name,
				},
			},
			Strategy: strategy,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: podLabels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: ms.name,
				},
			},
		},
	}
	containers := &dep.Spec.Template.Spec.Containers
	for idx := range ms.containers {
		msCont := &ms.containers[idx]
		cont := corev1.Container{
			Name:            msCont.name,
			Image:           msCont.image,
			ImagePullPolicy: corev1.PullPolicy(msCont.imagePullPolicy),
			Args:            msCont.args,
			ReadinessProbe:  msCont.readinessProbe,
			Ports:           msCont.ports,
			Env:             msCont.env,
			Command:         msCont.command,
		}
		*containers = append(*containers, cont)
	}
	return dep
}

func newServiceAccount(namespace string, ms *microservice) *corev1.ServiceAccount {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ms.name,
			Namespace: namespace,
		},
	}
	// If imagePullSecret is provided, add it to the ImagePullSecrets field
	if ms.imagePullSecret != "" {
		sa.ImagePullSecrets = []corev1.LocalObjectReference{
			{Name: ms.imagePullSecret},
		}
	}

	return sa
}

func newRoleBinding(namespace string, ms *microservice) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ms.name,
			Namespace: namespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind: "ServiceAccount",
				Name: ms.name,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     ms.name,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
}

func newRole(namespace string, ms *microservice) *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ms.name,
			Namespace: namespace,
		},
		Rules: ms.rbacRules,
	}
}
