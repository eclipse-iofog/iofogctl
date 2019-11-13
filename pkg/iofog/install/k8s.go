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

	crdapi "github.com/eclipse-iofog/iofog-operator/pkg/apis"
	iofogv1 "github.com/eclipse-iofog/iofog-operator/pkg/apis/iofog/v1"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	v1 "k8s.io/api/core/v1"
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

// Kubernetes struct to manage state of deployment on Kubernetes cluster
type Kubernetes struct {
	config          *restclient.Config
	kogClient       kogclient.Client
	clientset       *kubernetes.Clientset
	extsClientset   *extsclientset.Clientset
	ns              string
	ms              map[string]*microservice
	kogInstanceName string
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

	microservices := make(map[string]*microservice, 0)
	microservices["controller"] = &controllerMicroservice
	microservices["connector"] = &connectorMicroservice
	microservices["operator"] = &operatorMicroservice
	//microservices["scheduler"] = &schedulerMicroservice
	microservices["kubelet"] = &kubeletMicroservice

	return &Kubernetes{
		config:          config,
		clientset:       clientset,
		extsClientset:   extsClientset,
		ns:              namespace,
		ms:              microservices,
		kogInstanceName: "iokog",
	}, nil
}

func (k8s *Kubernetes) SetKubeletImage(image string) {
	if image != "" {
		k8s.ms["kubelet"].containers[0].image = image
	}
}

func (k8s *Kubernetes) SetOperatorImage(image string) {
	if image != "" {
		k8s.ms["operator"].containers[0].image = image
	}
}

func (k8s *Kubernetes) SetConnectorImage(image string) {
	if image != "" {
		k8s.ms["connector"].containers[0].image = image
	}
}

func (k8s *Kubernetes) SetControllerImage(image string) {
	if image != "" {
		k8s.ms["controller"].containers[0].image = image
	}
}

// CreateConnector on cluster
func (k8s *Kubernetes) CreateConnector(name string, user IofogUser) (err error) {
	if err := k8s.enableKogClient(); err != nil {
		return err
	}

	kogList := &iofogv1.KogList{}
	if err = k8s.kogClient.List(context.Background(), kogclient.InNamespace(k8s.ns), kogList); err != nil {
		return err
	}
	if len(kogList.Items) == 0 {
		return util.NewError("Could not find existing ioKog on the Kubernetes cluster")
	}
	var existingKog *iofogv1.Kog
	for _, kog := range kogList.Items {
		if kog.ObjectMeta.Name == k8s.kogInstanceName {
			existingKog = &kog
			break
		}
	}
	if existingKog == nil {
		return util.NewError("Could not find ioKog named " + k8s.kogInstanceName + " in namespace " + k8s.ns)
	}

	connectorExists := false
	for _, connector := range existingKog.Spec.Connectors.Instances {
		if connector.Name == name {
			connectorExists = true
			break
		}
	}
	if !connectorExists {
		existingKog.Spec.Connectors.Instances = append(existingKog.Spec.Connectors.Instances, iofogv1.Connector{
			Name: name,
		})
	}
	existingKog.Spec.Connectors.Image = k8s.ms["connector"].containers[0].image

	err = k8s.kogClient.Update(context.Background(), existingKog)
	if err != nil {
		return err
	}

	return nil
}

func (k8s *Kubernetes) enableCustomResources() error {
	// Kogs
	iokogCRD := newKogCRD()
	if _, err := k8s.extsClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(iokogCRD); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return err
		}
	}

	// Applications
	appCRD := newAppCRD()
	if _, err := k8s.extsClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(appCRD); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return err
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
	verbose("Creating namespace " + k8s.ns)
	ns := &v1.Namespace{
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
	verbose("Enabling CRDs")
	if err := k8s.enableCustomResources(); err != nil {
		return err
	}

	// Check if kog exists
	verbose("Finding existing Kog")
	kogKey := kogclient.ObjectKey{
		Name:      k8s.kogInstanceName,
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
				Name:      k8s.kogInstanceName,
				Namespace: k8s.ns,
			},
		}
	}

	// Encode credentials
	encodedUser := iofogv1.IofogUser{
		Email:    b64.StdEncoding.EncodeToString([]byte(user.Email)),
		Password: b64.StdEncoding.EncodeToString([]byte(user.Password)),
		Name:     user.Name,
		Surname:  user.Surname,
	}

	// Set specification
	kog.Spec = iofogv1.KogSpec{
		ControlPlane: iofogv1.ControlPlane{
			IofogUser:              encodedUser,
			ControllerReplicaCount: int32(replicas),
			ControllerImage:        k8s.ms["controller"].containers[0].image,
			KubeletImage:           k8s.ms["kubelet"].containers[0].image,
			Database:               iofogv1.Database(db),
			ServiceType:            k8s.ms["controller"].serviceType,
			LoadBalancerIP:         k8s.ms["controller"].IP,
		},
		Connectors: iofogv1.Connectors{
			Instances: []iofogv1.Connector{},
		},
	}

	// Create or update Kog
	if found {
		verbose("Updating existing Kog")
		if err := k8s.kogClient.Update(context.Background(), &kog); err != nil {
			return err
		}
	} else {
		verbose("Deploying new Kog")
		if err := k8s.kogClient.Create(context.Background(), &kog); err != nil {
			return err
		}
	}

	return nil
}

func (k8s *Kubernetes) deleteOperator() (err error) {
	// Resource name for deletions
	name := k8s.ms["operator"].name

	// Service Account
	if err = k8s.clientset.CoreV1().ServiceAccounts(k8s.ns).Delete(name, &metav1.DeleteOptions{}); err != nil {
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
	opSvcAcc := newServiceAccount(k8s.ns, k8s.ms["operator"])
	if _, err = k8s.clientset.CoreV1().ServiceAccounts(k8s.ns).Create(opSvcAcc); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return
		}
	}

	// Cluster Role Binding
	crb := newClusterRoleBinding(k8s.ns, k8s.ms["operator"])
	if _, err = k8s.clientset.RbacV1().ClusterRoleBindings().Create(crb); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return
		}
	}

	// Deployment
	opDep := newDeployment(k8s.ns, k8s.ms["operator"])
	if _, err = k8s.clientset.AppsV1().Deployments(k8s.ns).Create(opDep); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return
		}
		// Redeploy the operator
		if err = k8s.clientset.AppsV1().Deployments(k8s.ns).Delete(k8s.ms["operator"].name, &metav1.DeleteOptions{}); err != nil {
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
			Name:      k8s.kogInstanceName,
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

func (k8s *Kubernetes) DeleteConnector(name string) error {
	// Prepare kog client
	if err := k8s.enableKogClient(); err != nil {
		return err
	}

	// Find existing kog
	kogList := &iofogv1.KogList{}
	if err := k8s.kogClient.List(context.Background(), kogclient.InNamespace(k8s.ns), kogList); err != nil {
		return err
	}
	if len(kogList.Items) == 0 {
		return util.NewError("Could not find existing ioKog on the Kubernetes cluster")
	}
	var existingKog *iofogv1.Kog
	for _, kog := range kogList.Items {
		if kog.ObjectMeta.Name == k8s.kogInstanceName {
			existingKog = &kog
			break
		}
	}
	if existingKog == nil {
		return util.NewError("Could not find ioKog named " + k8s.kogInstanceName + " in namespace " + k8s.ns)
	}

	// Update existing kog
	for idx, connector := range existingKog.Spec.Connectors.Instances {
		if connector.Name == name {
			instances := existingKog.Spec.Connectors.Instances
			existingKog.Spec.Connectors.Instances = append(instances[:idx], instances[idx+1:]...)
			if err := k8s.kogClient.Update(context.Background(), existingKog); err != nil {
				return err
			}
			break
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
		svc, ok := event.Object.(*v1.Service)
		if !ok {
			err = util.NewInternalError("Failed to wait for services in namespace: " + k8s.ns)
			return
		}

		// Ignore irrelevant service events
		if svc.Name != name {
			continue
		}

		switch svc.Spec.Type {
		case "LoadBalancer":
			// Load balancer must be ready
			if len(svc.Status.LoadBalancer.Ingress) == 0 {
				continue
			}
			nodePort = targetPort
			ip = svc.Status.LoadBalancer.Ingress[0].IP

		case "NodePort":
			// Get a list of K8s nodes and return one of their external IPs
			var nodeList *v1.NodeList
			nodeList, err = k8s.clientset.CoreV1().Nodes().List(metav1.ListOptions{})
			if err == nil {
				if len(nodeList.Items) == 0 {
					err = util.NewError("Could not find Kubernetes nodes when waiting for NodePort service " + name)
				} else {
					// Return external IP of any of the nodes in the cluster
					for _, node := range nodeList.Items {
						for _, addrs := range node.Status.Addresses {
							if addrs.Type == "ExternalIP" {
								ip = addrs.Address
								break
							}
						}
					}
					if ip == "" {
						err = util.NewError("Could not get an external IP address of any Kubernetes nodes for NodePort service " + name)
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

		case "ClusterIP":
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
	if svcType != "LoadBalancer" && svcType != "NodePort" && svcType != "ClusterIP" {
		err = util.NewInputError("Tried to set K8s Controller Service type to " + svcType + ". Only LoadBalancer, NodePort, and ClusterIP types are acceptable.")
	}
	k8s.ms["controller"].serviceType = svcType
	return
}

func (k8s *Kubernetes) SetControllerIP(ip string) {
	if ip != "" {
		k8s.ms["controller"].IP = ip
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
		if svc.Name == k8s.ms["controller"].name {
			return nil
		}
	}
	return util.NewError("Could not find Controller Service in Kubernetes namespace " + namespace)
}

func (k8s *Kubernetes) GetControllerEndpoint() (endpoint string, err error) {
	ms := k8s.ms["controller"]
	ip, port, err := k8s.waitForService(ms.name, ms.ports[0])
	if err != nil {
		return
	}
	endpoint = fmt.Sprintf("%s:%d", ip, port)
	return
}

func (k8s *Kubernetes) GetConnectorEndpoint(name string) (endpoint string, err error) {
	ms := k8s.ms["connector"]
	// TODO: This name formatting is magic that depends on the operator
	ip, port, err := k8s.waitForService("connector-"+name, ms.ports[0])
	if err != nil {
		return
	}
	endpoint = fmt.Sprintf("%s:%d", ip, port)
	return
}
