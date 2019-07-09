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
	"fmt"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"k8s.io/api/core/v1"
	extsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"strings"
	"time"
)

// Kubernetes struct to manage state of deployment on Kubernetes cluster
type Kubernetes struct {
	configFilename string
	clientset      *kubernetes.Clientset
	extsClientset  *extsclientset.Clientset
	crdName        string
	ns             string
	ms             map[string]*microservice
}

// NewKubernetes constructs an object to manage cluster
func NewKubernetes(configFilename, namespace string) (*Kubernetes, error) {
	// Replace ~ in filename
	configFilename, err := util.ReplaceTilde(configFilename)
	if err != nil {
		return nil, err
	}

	// Get the kubernetes config from the filepath.
	config, err := clientcmd.BuildConfigFromFlags("", configFilename)
	if err != nil {
		return nil, err
	}

	// Instantiate Kubernetes client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	extsClientset, err := extsclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	microservices := make(map[string]*microservice, 0)
	microservices["controller"] = &controllerMicroservice
	microservices["connector"] = &connectorMicroservice
	microservices["operator"] = &operatorMicroservice
	microservices["scheduler"] = &schedulerMicroservice
	microservices["kubelet"] = &kubeletMicroservice

	return &Kubernetes{
		configFilename: configFilename,
		clientset:      clientset,
		extsClientset:  extsClientset,
		crdName:        "iofogs.k8s.iofog.org",
		ns:             namespace,
		ms:             microservices,
	}, nil
}

func (k8s *Kubernetes) SetImages(images map[string]string) error {
	for key, img := range images {
		if _, exists := k8s.ms[key]; !exists {
			return util.NewInputError("Invalid ioFog service image name specified: " + key)
		}
		k8s.ms[key].containers[0].image = img
	}
	return nil
}

func (k8s *Kubernetes) SetControllerIP(ip string) {
	k8s.ms["controller"].IP = ip
}

func (k8s *Kubernetes) GetControllerEndpoint() (endpoint string, err error) {
	// Check service exists
	doesNotExistMsg := "Kubernetes Service controller in namespace " + k8s.ns
	svcs, err := k8s.clientset.CoreV1().Services(k8s.ns).List(metav1.ListOptions{})
	if err != nil {
		return
	}
	if svcs == nil || len(svcs.Items) == 0 {
		err = util.NewNotFoundError(doesNotExistMsg)
		return
	}
	found := false
	for _, svc := range svcs.Items {
		if svc.Name == "controller" {
			found = true
			break
		}
	}
	if !found {
		err = util.NewNotFoundError(doesNotExistMsg)
		return
	}

	// Wait for IP
	ip, err := k8s.waitForService(k8s.ms["controller"].name)
	if err != nil {
		return
	}
	println("")
	endpoint = fmt.Sprintf("%s:%d", ip, k8s.ms["controller"].ports[0])
	return
}

// CreateController on cluster
func (k8s *Kubernetes) CreateController(user User) (endpoint string, err error) {
	// Install ioFog Core
	token, ips, err := k8s.createCore(user)
	if err != nil {
		return
	}
	endpoint = fmt.Sprintf("%s:%d", ips["controller"], k8s.ms["controller"].ports[0])

	// Install ioFog K8s Extensions
	if err = k8s.createExtension(token, ips); err != nil {
		return
	}

	return
}

// DeleteController from cluster
func (k8s *Kubernetes) DeleteController() error {
	defer util.SpinStop()
	// Delete Deployments
	util.SpinStart("Deleting deployments")
	deps, err := k8s.clientset.AppsV1().Deployments(k8s.ns).List(metav1.ListOptions{})
	if err != nil {
		if !isNotFound(err) {
			return err
		}
	}
	for _, dep := range deps.Items {
		if err = k8s.clientset.AppsV1().Deployments(k8s.ns).Delete(dep.Name, &metav1.DeleteOptions{}); err != nil {
			if !isNotFound(err) {
				return err
			}
		}
	}

	// Delete Services
	util.SpinStart("Deleting services")
	svcs, err := k8s.clientset.CoreV1().Services(k8s.ns).List(metav1.ListOptions{})
	if err != nil {
		if !isNotFound(err) {
			return err
		}
	}
	for _, svc := range svcs.Items {
		if err = k8s.clientset.CoreV1().Services(k8s.ns).Delete(svc.Name, &metav1.DeleteOptions{}); err != nil {
			if !isNotFound(err) {
				return err
			}
		}
	}

	// Delete Service Accounts
	util.SpinStart("Deleting service accounts")
	svcAccs, err := k8s.clientset.CoreV1().ServiceAccounts(k8s.ns).List(metav1.ListOptions{})
	if err != nil {
		if !isNotFound(err) {
			return err
		}
	}
	for _, acc := range svcAccs.Items {
		if err = k8s.clientset.CoreV1().ServiceAccounts(k8s.ns).Delete(acc.Name, &metav1.DeleteOptions{}); err != nil {
			if !isNotFound(err) {
				return err
			}
		}
	}

	// Delete Kubelet Cluster Role Binding
	util.SpinStart("Deleting roles and role bindings")
	if err = k8s.clientset.RbacV1().ClusterRoleBindings().Delete(k8s.ms["kubelet"].name, &metav1.DeleteOptions{}); err != nil {
		if !isNotFound(err) {
			return err
		}
	}

	// Delete Roles
	roles, err := k8s.clientset.RbacV1().Roles(k8s.ns).List(metav1.ListOptions{})
	if err != nil {
		if !isNotFound(err) {
			return err
		}
	}
	for _, role := range roles.Items {
		if err = k8s.clientset.RbacV1().Roles(k8s.ns).Delete(role.Name, &metav1.DeleteOptions{}); err != nil {
			if !isNotFound(err) {
				return err
			}
		}
	}

	// Delete Role Bindings
	roleBinds, err := k8s.clientset.RbacV1().RoleBindings(k8s.ns).List(metav1.ListOptions{})
	if err != nil {
		if !isNotFound(err) {
			return err
		}
	}
	for _, bind := range roleBinds.Items {
		if err = k8s.clientset.RbacV1().RoleBindings(k8s.ns).Delete(bind.Name, &metav1.DeleteOptions{}); err != nil {
			if !isNotFound(err) {
				return err
			}
		}
	}

	// Delete CRD
	util.SpinStart("Deleting CRDs")
	if err = k8s.extsClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(k8s.crdName, &metav1.DeleteOptions{}); err != nil {
		if !isNotFound(err) {
			return err
		}
	}

	// Delete Namespace
	util.SpinStart("Deleting namespace")
	if k8s.ns != "default" {
		if err = k8s.clientset.CoreV1().Namespaces().Delete(k8s.ns, &metav1.DeleteOptions{}); err != nil {
			if !isNotFound(err) {
				return err
			}
		}
		// Wait for namespace to be removed
		for {
			nsList, err := k8s.clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
			if err != nil {
				return err
			}

			nsDeleted := true
			for _, ns := range nsList.Items {
				if ns.Name == k8s.ns {
					nsDeleted = false
					time.Sleep(1000 * time.Millisecond)
					break
				}
			}
			if nsDeleted {
				break
			}
		}
	}

	return nil
}

func (k8s *Kubernetes) createCore(user User) (token string, ips map[string]string, err error) {
	defer util.SpinStop()
	// Create namespace
	util.SpinStart("Creating namespace ")
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: k8s.ns,
		},
	}
	if _, err = k8s.clientset.CoreV1().Namespaces().Create(ns); err != nil {
		if !isAlreadyExists(err) {
			return
		}
	}

	// Create Controller and Connector Services and Pods
	coreMs := []*microservice{
		k8s.ms["controller"],
		k8s.ms["connector"],
	}
	for _, ms := range coreMs {
		util.SpinStart("Deploying " + strings.Title(ms.name) + " pods")
		dep := newDeployment(k8s.ns, ms)
		if _, err = k8s.clientset.AppsV1().Deployments(k8s.ns).Create(dep); err != nil {
			if !isAlreadyExists(err) {
				return
			}
			// Delete existing
			if err = k8s.clientset.AppsV1().Deployments(k8s.ns).Delete(dep.Name, &metav1.DeleteOptions{}); err != nil {
				return
			}
			if err = k8s.waitForPodTerminate(dep.Name); err != nil {
				return
			}
			// Create new
			if _, err = k8s.clientset.AppsV1().Deployments(k8s.ns).Create(dep); err != nil {
				return
			}
			if err = k8s.waitForPod(dep.Name); err != nil {
				return
			}
		}
		util.SpinStart("Deploying " + strings.Title(ms.name) + " service")
		svc := newService(k8s.ns, ms)
		if _, err = k8s.clientset.CoreV1().Services(k8s.ns).Create(svc); err != nil {
			if !isAlreadyExists(err) {
				return
			}
			// Get the existing svc
			var existingSvc *v1.Service
			existingSvc, err = k8s.clientset.CoreV1().Services(k8s.ns).Get(svc.Name, metav1.GetOptions{})
			if err != nil {
				return
			}
			// If trying to allocate a new static IP, we must recreate the service
			if ms.IP != "" && ms.IP != existingSvc.Spec.LoadBalancerIP {
				// Delete existing
				if err = k8s.clientset.CoreV1().Services(k8s.ns).Delete(svc.Name, &metav1.DeleteOptions{}); err != nil {
					return
				}
				// Create new
				if _, err = k8s.clientset.CoreV1().Services(k8s.ns).Create(svc); err != nil {
					return
				}
				// Wait for completion
				if _, err = k8s.waitForService(svc.Name); err != nil {
					return
				}
			}
		}
		util.SpinStart("Deploying " + strings.Title(ms.name) + " service account")
		svcAcc := newServiceAccount(k8s.ns, ms)
		if _, err = k8s.clientset.CoreV1().ServiceAccounts(k8s.ns).Create(svcAcc); err != nil {
			if !isAlreadyExists(err) {
				return
			}
		}
	}

	// Wait for pods
	for _, ms := range coreMs {
		if err = k8s.waitForPod(ms.name); err != nil {
			return
		}
	}

	// Wait for services and get IPs
	ips = make(map[string]string)
	for _, ms := range coreMs {
		util.SpinStart("Waiting for " + strings.Title(ms.name) + " IP")
		var ip string
		ip, err = k8s.waitForService(ms.name)
		if err != nil {
			return
		}
		ips[ms.name] = ip
	}

	// Connect to controller
	endpoint := fmt.Sprintf("%s:%d", ips["controller"], k8s.ms["controller"].ports[0])
	ctrl := NewController(endpoint)

	// Create user (this is the first API call and the service might need to resolve IP to new pods so we retry)
	util.SpinStart("Waiting for Controller API and creating new user in Controller")
	connected := false
	for !connected {
		if err = ctrl.CreateUser(user); err != nil {
			if !strings.Contains(err.Error(), "already an account associated") {
				if strings.Contains(err.Error(), "connection refused") {
					continue
				}
				return
			}
		} else {
			connected = true
			continue
		}
		time.Sleep(time.Millisecond * 1000)
	}

	// Get token
	loginRequest := LoginRequest{
		Email:    user.Email,
		Password: user.Password,
	}
	loginResponse, err := ctrl.Login(loginRequest)
	if err != nil {
		return
	}
	token = loginResponse.AccessToken

	// Connect Controller with Connector
	util.SpinStart("Provisioning Connector to Controller")
	connectorRequest := ConnectorInfo{
		IP:      ips["connector"],
		DevMode: true,
		Domain:  ips["connector"],
		Name:    "gke",
	}
	if err = ctrl.AddConnector(connectorRequest, token); err != nil {
		return
	}

	err = nil
	return
}

func (k8s *Kubernetes) createExtension(token string, ips map[string]string) (err error) {
	defer util.SpinStop()

	// Create Scheduler resources
	util.SpinStart("Deploying Scheduler")
	schedDep := newDeployment(k8s.ns, k8s.ms["scheduler"])
	if _, err = k8s.clientset.AppsV1().Deployments(k8s.ns).Create(schedDep); err != nil {
		if !isAlreadyExists(err) {
			return
		}
		// Delete existing
		if err = k8s.clientset.AppsV1().Deployments(k8s.ns).Delete(schedDep.Name, &metav1.DeleteOptions{}); err != nil {
			return
		}
		if err = k8s.waitForPodTerminate(schedDep.Name); err != nil {
			return
		}
		// Create new
		if _, err = k8s.clientset.AppsV1().Deployments(k8s.ns).Create(schedDep); err != nil {
			return
		}
		if err = k8s.waitForPod(schedDep.Name); err != nil {
			return
		}
	}
	schedAcc := newServiceAccount(k8s.ns, k8s.ms["scheduler"])
	if _, err = k8s.clientset.CoreV1().ServiceAccounts(k8s.ns).Create(schedAcc); err != nil {
		if !isAlreadyExists(err) {
			return
		}
	}

	// Create Kubelet resources
	util.SpinStart("Deploying Kubelet")
	vkSvcAcc := newServiceAccount(k8s.ns, k8s.ms["kubelet"])
	if _, err = k8s.clientset.CoreV1().ServiceAccounts(k8s.ns).Create(vkSvcAcc); err != nil {
		if !isAlreadyExists(err) {
			return
		}
	}

	vkRoleBind := newClusterRoleBinding(k8s.ns, k8s.ms["kubelet"])
	if _, err = k8s.clientset.RbacV1().ClusterRoleBindings().Create(vkRoleBind); err != nil {
		if !isAlreadyExists(err) {
			return
		}
	}
	k8s.ms["kubelet"].containers[0].args = []string{
		"--namespace",
		k8s.ns,
		"--iofog-token",
		token,
		"--iofog-url",
		fmt.Sprintf("http://%s:%d", ips["controller"], k8s.ms["controller"].ports[0]),
	}
	vkDep := newDeployment(k8s.ns, k8s.ms["kubelet"])
	if _, err = k8s.clientset.AppsV1().Deployments(k8s.ns).Create(vkDep); err != nil {
		if !isAlreadyExists(err) {
			return
		}
		// Update it if it exists
		if _, err = k8s.clientset.AppsV1().Deployments(k8s.ns).Update(vkDep); err != nil {
			return
		}
	}

	// Create Operator resources
	util.SpinStart("Deploying Operator")
	opSvcAcc := newServiceAccount(k8s.ns, k8s.ms["operator"])
	if _, err = k8s.clientset.CoreV1().ServiceAccounts(k8s.ns).Create(opSvcAcc); err != nil {
		if !isAlreadyExists(err) {
			return
		}
	}
	opRole := newRole(k8s.ns, k8s.ms["operator"])
	if _, err = k8s.clientset.RbacV1().Roles(k8s.ns).Create(opRole); err != nil {
		if !isAlreadyExists(err) {
			return
		}
	}
	opRoleBind := newRoleBinding(k8s.ns, k8s.ms["operator"])
	if _, err = k8s.clientset.RbacV1().RoleBindings(k8s.ns).Create(opRoleBind); err != nil {
		if !isAlreadyExists(err) {
			return
		}
	}
	crd := newCustomResourceDefinition(k8s.crdName)
	if _, err = k8s.extsClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd); err != nil {
		if !isAlreadyExists(err) {
			return
		}
	}
	opDep := newDeployment(k8s.ns, k8s.ms["operator"])
	if _, err = k8s.clientset.AppsV1().Deployments(k8s.ns).Create(opDep); err != nil {
		if !isAlreadyExists(err) {
			return
		}
		// Update it if it exists
		if _, err = k8s.clientset.AppsV1().Deployments(k8s.ns).Update(opDep); err != nil {
			return
		}
	}

	err = nil
	return
}

func (k8s *Kubernetes) waitForPodTerminate(name string) error {
	terminating := false
	for !terminating {
		_, err := k8s.clientset.CoreV1().Pods(k8s.ns).Get(name, metav1.GetOptions{})
		if err != nil {
			terminating = strings.Contains(err.Error(), "not found")
			if !terminating {
				return err
			}
		}
		if !terminating {
			time.Sleep(time.Millisecond * 500)
		}
	}
	return nil
}

func (k8s *Kubernetes) waitForPod(name string) error {
	// Get watch handler to observe changes to pods
	watch, err := k8s.clientset.CoreV1().Pods(k8s.ns).Watch(metav1.ListOptions{})
	if err != nil {
		return err
	}

	// Wait for pod events
	for event := range watch.ResultChan() {
		// Get the pod
		pod, ok := event.Object.(*v1.Pod)
		if !ok {
			return util.NewInternalError("Failed to wait for pods in namespace: " + k8s.ns)
		}
		// Check pod is in running state
		if util.Before(pod.Name, "-") != name {
			continue
		}

		if pod.Status.Phase == "Running" {
			ready := true
			for _, cond := range pod.Status.Conditions {
				if cond.Status != "True" {
					ready = false
					break
				}
			}
			if ready {
				watch.Stop()
			}
		}
	}
	return nil
}

func (k8s *Kubernetes) waitForService(name string) (ip string, err error) {
	// Get watch handler to observe changes to services
	watch, err := k8s.clientset.CoreV1().Services(k8s.ns).Watch(metav1.ListOptions{})
	if err != nil {
		return
	}

	// Wait for Services to have IPs allocated
	for event := range watch.ResultChan() {
		svc, ok := event.Object.(*v1.Service)
		if !ok {
			err = util.NewInternalError("Failed to wait for services in namespace: " + k8s.ns)
			return
		}

		// Ignore irrelevant service events
		if svc.Name != name {
			continue
		}
		// Loadbalancer must be ready
		if len(svc.Status.LoadBalancer.Ingress) == 0 {
			continue
		}

		ip = svc.Status.LoadBalancer.Ingress[0].IP
		watch.Stop()
	}

	return
}

func isAlreadyExists(err error) bool {
	return strings.Contains(err.Error(), "already exists")
}

func isNotFound(err error) bool {
	return strings.Contains(err.Error(), "not found")
}
