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
	"fmt"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"k8s.io/api/core/v1"
	extsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"strconv"
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
	//microservices["scheduler"] = &schedulerMicroservice
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
		if img == "" {
			util.PrintNotify("Empty image name specified for " + key + ". Ignoring and using default")
			continue
		}
		if _, exists := k8s.ms[key]; !exists {
			return util.NewInputError("Invalid ioFog service image name specified: " + key)
		}
		k8s.ms[key].containers[0].image = img
	}
	return nil
}

func (k8s *Kubernetes) getEndpoint(ms *microservice) (endpoint string, err error) {
	if len(ms.ports) == 0 {
		err = util.NewError("Requested endpoint of Microservice on K8s cluster that does not have an external API")
		return
	}
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
		if svc.Name == ms.name {
			found = true
			break
		}
	}
	if !found {
		err = util.NewNotFoundError(doesNotExistMsg)
		return
	}

	// Wait for IP
	ip, err := k8s.waitForService(ms.name)
	if err != nil {
		return
	}
	endpoint = fmt.Sprintf("%s:%d", ip, ms.ports[0])
	return
}

// CreateConnector on cluster
func (k8s *Kubernetes) CreateConnector(name, controllerEndpoint string, user IofogUser) (err error) {
	// Install Connector
	if err = k8s.createDeploymentAndService(k8s.ms["connector"]); err != nil {
		return
	}
	// Get Connector endpoint
	connectorEndpoint, err := k8s.GetConnectorEndpoint()
	if err != nil {
		return
	}
	connectorIP := util.Before(connectorEndpoint, ":")

	// Log into Controller
	ctrlClient := client.New(controllerEndpoint)
	if err = ctrlClient.Login(client.LoginRequest{
		Email:    user.Email,
		Password: user.Password,
	}); err != nil {
		return
	}
	// Provision the Connector
	if err = ctrlClient.AddConnector(client.ConnectorInfo{
		IP:     connectorIP,
		Domain: connectorIP,
		Name:   name,
	}); err != nil {
		return err
	}

	return nil
}

// CreateController on cluster
func (k8s *Kubernetes) CreateController(replicas int) (err error) {
	// Configure replica count
	k8s.ms["controller"].replicas = int32(replicas)
	// Install Controller
	if err = k8s.createDeploymentAndService(k8s.ms["controller"]); err != nil {
		return
	}
	// Wait for Controller API
	verbose("Waiting for Controller API")
	endpoint, err := k8s.GetControllerEndpoint()
	if err = waitForControllerAPI(endpoint); err != nil {
		return
	}

	return
}
func (k8s *Kubernetes) DeleteAll() error {
	return k8s.delete(true)
}

func (k8s *Kubernetes) DeleteController() error {
	return k8s.delete(false)
}

func (k8s *Kubernetes) DeleteConnector() error {
	// Delete deployment
	if err := k8s.clientset.AppsV1().Deployments(k8s.ns).Delete("connector", &metav1.DeleteOptions{}); err != nil {
		if !isNotFound(err) {
			return err
		}
	}
	// Delete service
	if err := k8s.clientset.CoreV1().Services(k8s.ns).Delete("connector", &metav1.DeleteOptions{}); err != nil {
		if !isNotFound(err) {
			return err
		}
	}

	return nil
}

// DeleteController from cluster
func (k8s *Kubernetes) delete(all bool) error {
	// Delete Deployments
	deps, err := k8s.clientset.AppsV1().Deployments(k8s.ns).List(metav1.ListOptions{})
	if err != nil {
		if !isNotFound(err) {
			return err
		}
	}
	for _, dep := range deps.Items {
		if all || dep.Name != "connector" {
			if err = k8s.clientset.AppsV1().Deployments(k8s.ns).Delete(dep.Name, &metav1.DeleteOptions{}); err != nil {
				if !isNotFound(err) {
					return err
				}
			}
		}
	}

	// Delete Services
	svcs, err := k8s.clientset.CoreV1().Services(k8s.ns).List(metav1.ListOptions{})
	if err != nil {
		if !isNotFound(err) {
			return err
		}
	}
	for _, svc := range svcs.Items {
		if all || svc.Name != "connector" {
			if err = k8s.clientset.CoreV1().Services(k8s.ns).Delete(svc.Name, &metav1.DeleteOptions{}); err != nil {
				if !isNotFound(err) {
					return err
				}
			}
		}
	}

	// Delete Service Accounts
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
	if err = k8s.extsClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(k8s.crdName, &metav1.DeleteOptions{}); err != nil {
		if !isNotFound(err) {
			return err
		}
	}

	// Delete Namespace
	if k8s.ns != "default" && all {
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

func (k8s *Kubernetes) createDeploymentAndService(ms *microservice) (err error) {
	// Create namespace
	verbose("Creating namespace")
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
	verbose("Creating " + ms.name + " Deployment and Service")
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
	svcAcc := newServiceAccount(k8s.ns, ms)
	if _, err = k8s.clientset.CoreV1().ServiceAccounts(k8s.ns).Create(svcAcc); err != nil {
		if !isAlreadyExists(err) {
			return
		}
	}

	// Wait for pods
	verbose("Waiting for " + ms.name + " Pods")
	if err = k8s.waitForPod(ms.name); err != nil {
		return
	}

	// Wait for services and get IPs
	verbose("Waiting for Service IPs")
	_, err = k8s.waitForService(ms.name)
	if err != nil {
		return
	}

	return
}

func (k8s *Kubernetes) CreateExtensionServices(user IofogUser) (err error) {
	// Login in and retrieve access token for Kubelet
	verbose("Logging into Controller")
	endpoint, err := k8s.GetControllerEndpoint()
	if err != nil {
		return
	}
	ctrlClient := client.New(endpoint)
	if err = ctrlClient.Login(client.LoginRequest{
		Email:    user.Email,
		Password: user.Password,
	}); err != nil {
		return
	}
	token := ctrlClient.GetAccessToken()

	verbose("Deploying Operator and Kubelet")
	// Create Scheduler resources
	//schedDep := newDeployment(k8s.ns, k8s.ms["scheduler"])
	//if _, err = k8s.clientset.AppsV1().Deployments(k8s.ns).Create(schedDep); err != nil {
	//	if !isAlreadyExists(err) {
	//		return
	//	}
	//	// Delete existing
	//	if err = k8s.clientset.AppsV1().Deployments(k8s.ns).Delete(schedDep.Name, &metav1.DeleteOptions{}); err != nil {
	//		return
	//	}
	//	if err = k8s.waitForPodTerminate(schedDep.Name); err != nil {
	//		return
	//	}
	//	// Create new
	//	if _, err = k8s.clientset.AppsV1().Deployments(k8s.ns).Create(schedDep); err != nil {
	//		return
	//	}
	//	if err = k8s.waitForPod(schedDep.Name); err != nil {
	//		return
	//	}
	//}
	//schedAcc := newServiceAccount(k8s.ns, k8s.ms["scheduler"])
	//if _, err = k8s.clientset.CoreV1().ServiceAccounts(k8s.ns).Create(schedAcc); err != nil {
	//	if !isAlreadyExists(err) {
	//		return
	//	}
	//}

	// Create Kubelet resources
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
		fmt.Sprintf("http://%s:%d", "controller", k8s.ms["controller"].ports[0]),
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

func (k8s *Kubernetes) SetControllerExternalDatabase(host, user, password string, port int) {
	k8s.ms["controller"].containers[0].env = []v1.EnvVar{
		{
			Name:  "DB_PROVIDER",
			Value: "postgres",
		},
		{
			Name:  "DB_USERNAME",
			Value: user,
		},
		{
			Name:  "DB_PASSWORD",
			Value: password,
		},
		{
			Name:  "DB_HOST",
			Value: host,
		},
		{
			Name:  "DB_PORT",
			Value: strconv.Itoa(port),
		},
	}
}

func (k8s *Kubernetes) SetControllerIP(ip string) {
	k8s.ms["controller"].IP = ip
}

func (k8s *Kubernetes) GetControllerEndpoint() (endpoint string, err error) {
	return k8s.getEndpoint(k8s.ms["controller"])
}

func (k8s *Kubernetes) GetConnectorEndpoint() (endpoint string, err error) {
	return k8s.getEndpoint(k8s.ms["connector"])
}
