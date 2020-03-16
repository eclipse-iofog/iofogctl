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
	"context"
	"fmt"

	ioclient "github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	crdapi "github.com/eclipse-iofog/iofog-operator/v2/pkg/apis"
	iofogv1 "github.com/eclipse-iofog/iofog-operator/v2/pkg/apis/iofog"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
	corev1 "k8s.io/api/core/v1"
	extsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	extsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	runtime "k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	kogclient "sigs.k8s.io/controller-runtime/pkg/client"

	b64 "encoding/base64"
)

const (
	kogInstanceName = "iokog"
	dockerRepo      = "iofog"
)

// Kubernetes struct to manage state of deployment on Kubernetes cluster
type Kubernetes struct {
	config        *restclient.Config
	kogClient     kogclient.Client
	clientset     *kubernetes.Clientset
	extsClientset *extsclientset.Clientset
	ns            string
	operator      *microservice
	controlPlane  *iofogv1.ControlPlane
}

// NewKubernetes constructs an object to manage cluster
func NewKubernetes(configFilename, namespace string) (*Kubernetes, error) {
	// Get the kubernetes config from the filepath.
	config, err := clientcmd.BuildConfigFromFlags("", configFilename)
	if err != nil {
		return nil, err
	}

	// Instantiate Kubernetes clients
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	extsClientset, err := extsclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Kubernetes{
		config:        config,
		clientset:     clientset,
		extsClientset: extsClientset,
		ns:            namespace,
		controlPlane: &iofogv1.ControlPlane{
			ControllerImage:  dockerRepo + "/controller:" + util.GetControllerTag(),
			PortManagerImage: dockerRepo + "/port-manager:" + util.GetPortManagerTag(),
			RouterImage:      dockerRepo + "/router:" + util.GetRouterTag(),
			ProxyImage:       dockerRepo + "/proxy:" + util.GetProxyTag(),
			KubeletImage:     dockerRepo + "/kubelet:" + util.GetKubeletTag(),
			ServiceType:      string(corev1.ServiceTypeLoadBalancer),
		},
		operator: newOperatorMicroservice(),
	}, nil
}

func (k8s *Kubernetes) SetKubeletImage(image string) {
	if image != "" {
		k8s.controlPlane.KubeletImage = image
	}
}

func (k8s *Kubernetes) SetOperatorImage(image string) {
	if image != "" {
		k8s.operator.containers[0].image = image
	}
}

func (k8s *Kubernetes) SetPortManagerImage(image string) {
	if image != "" {
		k8s.controlPlane.PortManagerImage = image
	}
}

func (k8s *Kubernetes) SetRouterImage(image string) {
	if image != "" {
		k8s.controlPlane.RouterImage = image
	}
}

func (k8s *Kubernetes) SetProxyImage(image string) {
	if image != "" {
		k8s.controlPlane.ProxyImage = image
	}
}

func (k8s *Kubernetes) SetControllerImage(image string) {
	if image != "" {
		k8s.controlPlane.ControllerImage = image
	}
}

func (k8s *Kubernetes) enableCustomResources() error {
	// Kog and App
	for _, crd := range []*extsv1.CustomResourceDefinition{newKogCRD(), newAppCRD()} {
		// Try create new
		if _, err := k8s.extsClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd); err != nil {
			if !k8serrors.IsAlreadyExists(err) {
				return err
			}
			// Update
			if _, err := k8s.extsClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Update(crd); err != nil {
				return err
			}
		}
	}

	// Deploy operator again
	if err := k8s.createOperator(); err != nil {
		return err
	}

	// Enable client after CRDs have been made
	if err := k8s.enableKogClient(); err != nil {
		return err
	}

	return nil
}

func (k8s *Kubernetes) enableKogClient() (err error) {
	scheme := runtime.NewScheme()
	clientgoscheme.AddToScheme(scheme)
	crdapi.AddToScheme(scheme)
	k8s.kogClient, err = kogclient.New(k8s.config, kogclient.Options{Scheme: scheme})
	if err != nil {
		return err
	}
	return nil
}

// CreateController on cluster
func (k8s *Kubernetes) CreateController(user IofogUser, replicas int, db Database) error {
	// Create namespace if required
	Verbose("Creating namespace " + k8s.ns)
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: k8s.ns,
		},
	}
	if _, err := k8s.clientset.CoreV1().Namespaces().Create(ns); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return err
		}
	}

	// Set up CRDs if required
	Verbose("Enabling CRDs")
	if err := k8s.enableCustomResources(); err != nil {
		return err
	}

	// Check if kog exists
	Verbose("Finding existing Kog")
	kogKey := kogclient.ObjectKey{
		Name:      kogInstanceName,
		Namespace: k8s.ns,
	}
	var kog iofogv1.Kog
	found := true
	if err := k8s.kogClient.Get(context.Background(), kogKey, &kog); err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
		// Not found, set basic info
		found = false
		kog = iofogv1.Kog{
			ObjectMeta: metav1.ObjectMeta{
				Name:      kogInstanceName,
				Namespace: k8s.ns,
			},
		}
	}

	// Encode credentials
	user.Password = b64.StdEncoding.EncodeToString([]byte(user.Password))

	// Set specification
	kog.Spec = iofogv1.KogSpec{
		ControlPlane: *k8s.controlPlane,
	}
	kog.Spec.ControlPlane.ControllerReplicaCount = int32(replicas)
	kog.Spec.ControlPlane.Database = iofogv1.Database(db)
	kog.Spec.ControlPlane.IofogUser = iofogv1.IofogUser(user)

	// Create or update Kog
	if found {
		Verbose("Updating existing Kog")
		if err := k8s.kogClient.Update(context.Background(), &kog); err != nil {
			return err
		}
	} else {
		Verbose("Deploying new Kog")
		if err := k8s.kogClient.Create(context.Background(), &kog); err != nil {
			return err
		}
	}

	return nil
}

func (k8s *Kubernetes) deleteOperator() (err error) {
	// Resource name for deletions
	name := k8s.operator.name

	// Service Account
	if err = k8s.clientset.CoreV1().ServiceAccounts(k8s.ns).Delete(name, &metav1.DeleteOptions{}); err != nil {
		if !k8serrors.IsNotFound(err) {
			return
		}
	}

	// Role
	if err = k8s.clientset.RbacV1().Roles(k8s.ns).Delete(name, &metav1.DeleteOptions{}); err != nil {
		if !k8serrors.IsNotFound(err) {
			return
		}
	}

	// Role Binding
	if err = k8s.clientset.RbacV1().RoleBindings(k8s.ns).Delete(name, &metav1.DeleteOptions{}); err != nil {
		if !k8serrors.IsNotFound(err) {
			return
		}
	}

	// Cluster Role Binding
	if err = k8s.clientset.RbacV1().ClusterRoleBindings().Delete(getClusterRoleBindingName(k8s.ns, name), &metav1.DeleteOptions{}); err != nil {
		if !k8serrors.IsNotFound(err) {
			return
		}
	}

	// Deployment
	if err = k8s.clientset.AppsV1().Deployments(k8s.ns).Delete(name, &metav1.DeleteOptions{}); err != nil {
		if !k8serrors.IsNotFound(err) {
			return
		}
	}

	return
}

func (k8s *Kubernetes) createOperator() (err error) {
	// Service Account
	opSvcAcc := newServiceAccount(k8s.ns, k8s.operator)
	if _, err = k8s.clientset.CoreV1().ServiceAccounts(k8s.ns).Create(opSvcAcc); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return
		}
	}

	// Role
	role := newRole(k8s.ns, k8s.operator)
	if _, err = k8s.clientset.RbacV1().Roles(k8s.ns).Create(role); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return
		}
	}

	// Role Binding
	rb := newRoleBinding(k8s.ns, k8s.operator)
	if _, err = k8s.clientset.RbacV1().RoleBindings(k8s.ns).Create(rb); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return
		}
	}

	// Cluster Role Binding
	crb := newClusterRoleBinding(k8s.ns, k8s.operator)
	if _, err = k8s.clientset.RbacV1().ClusterRoleBindings().Create(crb); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return
		}
	}

	// Deployment
	opDep := newDeployment(k8s.ns, k8s.operator)
	if _, err = k8s.clientset.AppsV1().Deployments(k8s.ns).Create(opDep); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return
		}
		// Redeploy the operator
		if err = k8s.clientset.AppsV1().Deployments(k8s.ns).Delete(k8s.operator.name, &metav1.DeleteOptions{}); err != nil {
			return
		}
		if _, err = k8s.clientset.AppsV1().Deployments(k8s.ns).Create(opDep); err != nil {
			return
		}
	}
	return nil
}

func (k8s *Kubernetes) DeleteController() error {
	// Prepare kog client
	if err := k8s.enableKogClient(); err != nil {
		return err
	}

	// Delete Kog
	kog := &iofogv1.Kog{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kogInstanceName,
			Namespace: k8s.ns,
		},
	}
	if err := k8s.kogClient.Delete(context.Background(), kog); err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
	}

	// Delete Operator
	if err := k8s.deleteOperator(); err != nil {
		return err
	}

	// Delete Namespace
	if k8s.ns != "default" {
		if err := k8s.clientset.CoreV1().Namespaces().Delete(k8s.ns, &metav1.DeleteOptions{}); err != nil {
			if !k8serrors.IsNotFound(err) {
				return err
			}
		}
	}

	return nil
}

func (k8s *Kubernetes) waitForService(name string, targetPort int32) (ip string, nodePort int32, err error) {
	// Get watch handler to observe changes to services
	watch, err := k8s.clientset.CoreV1().Services(k8s.ns).Watch(metav1.ListOptions{})
	if err != nil {
		return
	}

	// Wait for Services to have IPs allocated
	for event := range watch.ResultChan() {
		svc, ok := event.Object.(*corev1.Service)
		if !ok {
			err = util.NewInternalError("Failed to wait for services in namespace: " + k8s.ns)
			return
		}

		// Ignore irrelevant service events
		if svc.Name != name {
			continue
		}

		switch svc.Spec.Type {
		case corev1.ServiceTypeLoadBalancer:
			// Load balancer must be ready
			if len(svc.Status.LoadBalancer.Ingress) == 0 {
				continue
			}
			nodePort = targetPort
			ip = svc.Status.LoadBalancer.Ingress[0].IP

		case corev1.ServiceTypeNodePort:
			// Get a list of K8s nodes and return one of their external IPs
			var nodeList *corev1.NodeList
			nodeList, err = k8s.clientset.CoreV1().Nodes().List(metav1.ListOptions{})
			if err == nil {
				if len(nodeList.Items) == 0 {
					err = util.NewError("Could not find Kubernetes nodes when waiting for NodePort service " + name)
				} else {
					// Return external IP of any of the nodes in the cluster
					for _, node := range nodeList.Items {
						for _, addrs := range node.Status.Addresses {
							if addrs.Type == corev1.NodeExternalIP {
								ip = addrs.Address
								break
							}
						}
					}
					if ip == "" {
						util.PrintNotify("Could not get an external IP address of any Kubernetes nodes for NodePort service " + name + "\nTrying to reach the cluster IP of the service")
						for _, node := range nodeList.Items {
							for _, addrs := range node.Status.Addresses {
								if addrs.Type == corev1.NodeInternalIP {
									ip = addrs.Address
									break
								}
							}
						}
						if ip == "" {
							err = util.NewError("Could not get an external or internal IP address of any Kubernetes nodes for NodePort service " + name)
						}
					}
				}
			}
			// Get the port allocated on the node
			if err == nil {
				for _, port := range svc.Spec.Ports {
					if port.TargetPort.IntVal == targetPort {
						nodePort = port.NodePort
						break
					}
				}
				if nodePort == 0 {
					err = util.NewError("Could not get node port for Kubernetes service " + name)
				}
			}

		case corev1.ServiceTypeClusterIP:
			// Note: ClusterIPs are internal to K8s cluster only
		}

		// End the loop
		watch.Stop()
	}

	return
}

func (k8s *Kubernetes) SetControllerServiceType(svcType string) (err error) {
	if svcType == "" {
		return nil
	}

	coreSvcType := corev1.ServiceType(svcType)
	if coreSvcType != corev1.ServiceTypeLoadBalancer && coreSvcType != corev1.ServiceTypeNodePort && coreSvcType != corev1.ServiceTypeClusterIP {
		err = util.NewInputError("Tried to set K8s Service type to " + svcType + ". Only LoadBalancer, NodePort, and ClusterIP types are acceptable.")
	}
	k8s.controlPlane.ServiceType = svcType
	return
}

func (k8s *Kubernetes) SetControllerIP(ip string) {
	if ip != "" {
		k8s.controlPlane.LoadBalancerIP = ip
	}
}

func (k8s *Kubernetes) ExistsInNamespace(namespace string) error {
	// Check namespace exists
	if _, err := k8s.clientset.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			return util.NewError("Could not find Namespace " + namespace + " on Kubernetes cluster")
		}
		return err
	}

	// Check services exist
	svcList, err := k8s.clientset.CoreV1().Services(namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, svc := range svcList.Items {
		if svc.Name == controller {
			return nil
		}
	}
	return util.NewError("Could not find Controller Service in Kubernetes namespace " + namespace)
}

func (k8s *Kubernetes) GetControllerEndpoint() (endpoint string, err error) {
	ip, port, err := k8s.waitForService(controller, ioclient.ControllerPort)
	if err != nil {
		return
	}
	return util.GetControllerEndpoint(fmt.Sprintf("%s:%d", ip, port))
}
