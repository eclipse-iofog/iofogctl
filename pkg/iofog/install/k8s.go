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
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	ioclient "github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	crdapi "github.com/eclipse-iofog/iofog-operator/v2/pkg/apis"
	iofogv2 "github.com/eclipse-iofog/iofog-operator/v2/pkg/apis/iofog"
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
	opclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	cpInstanceName = "iofog"
)

// Kubernetes struct to manage state of deployment on Kubernetes cluster
type Kubernetes struct {
	config        *restclient.Config
	opClient      opclient.Client
	clientset     *kubernetes.Clientset
	extsClientset *extsclientset.Clientset
	ns            string
	operator      *microservice
	services      iofogv2.Services
	images        iofogv2.Images
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
		operator:      newOperatorMicroservice(),
	}, nil
}

func (k8s *Kubernetes) SetKubeletImage(image string) {
	if image != "" {
		k8s.images.Kubelet = image
	} else {
		k8s.images.Kubelet = util.GetKubeletImage()
	}
}

func (k8s *Kubernetes) SetOperatorImage(image string) {
	if image != "" {
		k8s.operator.containers[0].image = image
	} else {
		k8s.operator.containers[0].image = util.GetOperatorImage()
	}
}

func (k8s *Kubernetes) SetPortManagerImage(image string) {
	if image != "" {
		k8s.images.PortManager = image
	} else {
		k8s.images.PortManager = util.GetPortManagerImage()
	}
}

func (k8s *Kubernetes) SetRouterImage(image string) {
	if image != "" {
		k8s.images.Router = image
	} else {
		k8s.images.Router = util.GetRouterImage()
	}
}

func (k8s *Kubernetes) SetProxyImage(image string) {
	if image != "" {
		k8s.images.Proxy = image
	} else {
		k8s.images.Proxy = util.GetProxyImage()
	}
}

func (k8s *Kubernetes) SetControllerImage(image string) {
	if image != "" {
		k8s.images.Controller = image
	} else {
		k8s.images.Controller = util.GetControllerImage()
	}
}

func (k8s *Kubernetes) enableCustomResources() error {
	// Control Plane and App
	for _, crd := range []*extsv1.CustomResourceDefinition{iofogv2.NewControlPlaneCustomResource(), iofogv2.NewAppCustomResource()} {
		// Try create new
		if _, err := k8s.extsClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd); err != nil {
			if !k8serrors.IsAlreadyExists(err) {
				return err
			}
			// Update
			existingCRD, err := k8s.extsClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get(crd.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			if !iofogv2.IsSupportedCustomResource(existingCRD) {
				existingCRD.Spec.Versions = crd.Spec.Versions
				if _, err := k8s.extsClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Update(existingCRD); err != nil {
					return err
				}
			}
		}
	}

	// Deploy operator again
	if err := k8s.createOperator(); err != nil {
		return err
	}

	// Enable client after CRDs have been made
	if err := k8s.enableOperatorClient(); err != nil {
		return err
	}

	return nil
}

func (k8s *Kubernetes) enableOperatorClient() (err error) {
	scheme := runtime.NewScheme()
	clientgoscheme.AddToScheme(scheme)
	crdapi.AddToScheme(scheme)
	k8s.opClient, err = opclient.New(k8s.config, opclient.Options{Scheme: scheme})
	if err != nil {
		return err
	}
	return nil
}

// CreateController on cluster
func (k8s *Kubernetes) CreateController(user IofogUser, replicas int32, db Database) (endpoint string, err error) {
	// Create namespace if required
	Verbose("Creating namespace " + k8s.ns)
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: k8s.ns,
		},
	}
	if _, err = k8s.clientset.CoreV1().Namespaces().Create(ns); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return
		}
	}

	// Set up CRDs if required
	Verbose("Enabling CRDs")
	if err = k8s.enableCustomResources(); err != nil {
		return
	}

	// Check if Control Plane exists
	Verbose("Finding existing Control Plane")
	cpKey := opclient.ObjectKey{
		Name:      cpInstanceName,
		Namespace: k8s.ns,
	}
	var cp iofogv2.ControlPlane
	found := true
	if err = k8s.opClient.Get(context.Background(), cpKey, &cp); err != nil {
		if !k8serrors.IsNotFound(err) {
			return
		}
		// Not found, set basic info
		found = false
		cp = iofogv2.ControlPlane{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cpInstanceName,
				Namespace: k8s.ns,
			},
		}
	}

	// Set specification
	cp.Spec.Replicas.Controller = int32(replicas)
	cp.Spec.Database = iofogv2.Database(db)
	cp.Spec.User = iofogv2.User(user)
	cp.Spec.Services = k8s.services
	cp.Spec.Images = k8s.images

	// Create or update Control Plane
	if found {
		Verbose("Updating existing Control Plane")
		if err = k8s.opClient.Update(context.Background(), &cp); err != nil {
			return
		}
	} else {
		Verbose("Deploying new Control Plane")
		if err = k8s.opClient.Create(context.Background(), &cp); err != nil {
			return
		}
	}

	// Get endpoint of deployed Controller
	endpoint, err = k8s.GetControllerEndpoint()
	if err != nil {
		return
	}

	// Wait for Default Router to be registered by Port Manager
	errCh := make(chan error, 1)
	go k8s.monitorOperator(errCh)
	select {
	case err = <-errCh:
	case <-time.After(240 * time.Second):
		err = util.NewInternalError("Failed to wait for Default Router registration")
	}

	return
}

// Watch Operator logs
// Report error from Operator if found in logs
// Operator Pods are deleted and created when Control Plane redeployed
func (k8s *Kubernetes) monitorOperator(errCh chan error) {
	defer close(errCh)
	errSuffix := "while awaiting finalization of Control Plane"
	for {
		time.Sleep(2 * time.Second)
		// Check operator logs
		pods, err := k8s.clientset.CoreV1().Pods(k8s.ns).List(metav1.ListOptions{
			LabelSelector: "name=iofog-operator", // TODO: Decouple this
		})
		if err != nil {
			errCh <- err
			return
		}
		if len(pods.Items) == 0 {
			errCh <- util.NewInternalError("Could not find any Operator Pods " + errSuffix)
			return
		}
		// Find ready Pod
		var pod *corev1.Pod
		for podIdx := range pods.Items {
			for _, condition := range pods.Items[podIdx].Status.Conditions {
				if condition.Type == corev1.PodReady {
					if condition.Status == corev1.ConditionTrue {
						pod = &pods.Items[podIdx]
						break
					}
				}
			}
			if pod != nil {
				break
			}
		}
		// Could not find ready Operator Pod
		if pod == nil {
			continue
		}
		// Get the logs of ready Pod
		req := k8s.clientset.CoreV1().Pods(k8s.ns).GetLogs(pod.Name, &corev1.PodLogOptions{})
		podLogs, err := req.Stream()
		if err != nil {
			errCh <- util.NewInternalError("Error opening Operator Pod log stream " + errSuffix)
			return
		}
		defer podLogs.Close()
		buf := new(bytes.Buffer)
		if _, err = io.Copy(buf, podLogs); err != nil {
			errCh <- util.NewInternalError("Error reading Operator Pod log stream " + errSuffix)
			return
		}
		podLogsStr := buf.String()
		if strings.Contains(podLogsStr, `"msg":"Completed Reconciliation","Request.Namespace":"`+k8s.ns) { // TODO: Decouple iofogctl-operator succ string
			errCh <- nil
			return
		}
		errDelim := `"level":"error"` // TODO: Decouple iofogctl-operator err string
		if strings.Contains(podLogsStr, errDelim) {
			msg := ""
			logLines := strings.Split(podLogsStr, "\n")
			for _, line := range logLines {
				if strings.Contains(line, errDelim) {
					msg = fmt.Sprintf("%s\n%s", msg, line)
				}
			}
			errCh <- util.NewInternalError("Operator failed to reconcile Control Plane " + msg)
			return
		}

		// Continue loop, wait for Router registration or error...
	}
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
	// Prepare Control Plane client
	if err := k8s.enableOperatorClient(); err != nil {
		return err
	}

	// Delete Control Plane
	cp := &iofogv2.ControlPlane{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cpInstanceName,
			Namespace: k8s.ns,
		},
	}
	if err := k8s.opClient.Delete(context.Background(), cp); err != nil {
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

func (k8s *Kubernetes) waitForService(name string, targetPort int32) (addr string, nodePort int32, err error) {
	// Get watch handler to observe changes to services
	watch, err := k8s.clientset.CoreV1().Services(k8s.ns).Watch(metav1.ListOptions{})
	if err != nil {
		return
	}

	// Wait for Services to have addresses allocated
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
			ip := svc.Status.LoadBalancer.Ingress[0].IP
			host := svc.Status.LoadBalancer.Ingress[0].Hostname
			if ip != "" {
				addr = ip
			}
			if host != "" {
				addr = host
			}
			if addr == "" {
				continue
			}

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
								addr = addrs.Address
								break
							}
						}
					}
					if addr == "" {
						util.PrintNotify("Could not get an external IP address of any Kubernetes nodes for NodePort service " + name + "\nTrying to reach the cluster IP of the service")
						for _, node := range nodeList.Items {
							for _, addrs := range node.Status.Addresses {
								if addrs.Type == corev1.NodeInternalIP {
									addr = addrs.Address
									break
								}
							}
						}
						if addr == "" {
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

func (k8s *Kubernetes) SetControllerService(svcType, ip string) {
	if svcType != "" {
		k8s.services.Controller.Type = svcType
	} else {
		k8s.services.Controller.Type = string(corev1.ServiceTypeLoadBalancer)
	}
	k8s.services.Controller.Address = ip
}

func (k8s *Kubernetes) SetRouterService(svcType, ip string) {
	if svcType != "" {
		k8s.services.Router.Type = svcType
	} else {
		k8s.services.Router.Type = string(corev1.ServiceTypeLoadBalancer)
	}
	k8s.services.Router.Address = ip
}

func (k8s *Kubernetes) SetProxyService(svcType, ip string) {
	if svcType != "" {
		k8s.services.Proxy.Type = svcType
	} else {
		k8s.services.Proxy.Type = string(corev1.ServiceTypeLoadBalancer)
	}
	k8s.services.Proxy.Address = ip
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

func (k8s *Kubernetes) GetControllerPods() (podNames []Pod, err error) {
	podNames = make([]Pod, 0)
	// List pods
	pods, err := k8s.clientset.CoreV1().Pods(k8s.ns).List(metav1.ListOptions{})
	if err != nil {
		return
	}
	// Find Controller pods
	for idx := range pods.Items {
		if pods.Items[idx].Labels["name"] == controller {
			podNames = append(podNames, Pod{
				Name:   pods.Items[idx].Name,
				Status: string(pods.Items[idx].Status.Phase),
			})
		}
	}
	return
}
