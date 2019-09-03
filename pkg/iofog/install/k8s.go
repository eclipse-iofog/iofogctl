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
	"github.com/eclipse-iofog/iofog-operator/pkg/apis/k8s/v1alpha2"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"k8s.io/api/core/v1"
	extsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"strconv"
	"strings"
	"time"

	runtime "k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	k8sconfig "sigs.k8s.io/controller-runtime/pkg/client/config"
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
	//// Check service exists
	//doesNotExistMsg := "Kubernetes Service " + ms.name + " in namespace " + k8s.ns
	//svcs, err := k8s.clientset.CoreV1().Services(k8s.ns).List(metav1.ListOptions{})
	//if err != nil {
	//	return
	//}
	//if svcs == nil || len(svcs.Items) == 0 {
	//	err = util.NewNotFoundError(doesNotExistMsg)
	//	return
	//}
	//found := false
	//for _, svc := range svcs.Items {
	//	if svc.Name == ms.name {
	//		found = true
	//		break
	//	}
	//}
	//if !found {
	//	err = util.NewNotFoundError(doesNotExistMsg)
	//	return
	//}

	// Wait for IP
	ip, err := k8s.waitForService(ms.name)
	if err != nil {
		return
	}
	endpoint = fmt.Sprintf("%s:%d", ip, ms.ports[0])
	return
}

// CreateConnector on cluster
func (k8s *Kubernetes) CreateConnector(name string, user IofogUser) (err error) {
	scheme := runtime.NewScheme()
	clientgoscheme.AddToScheme(scheme)
	crdapi.AddToScheme(scheme)

	cl, err := k8sclient.New(k8sconfig.GetConfigOrDie(), k8sclient.Options{Scheme: scheme})
	if err != nil {
		return err
	}

	kogList := &v1alpha2.KogList{}
	if err = cl.List(context.Background(), k8sclient.InNamespace(k8s.ns), kogList); err != nil {
		return err
	}
	if len(kogList.Items) == 0 {
		return util.NewError("Could not find existing ioKog on the Kubernetes cluster")
	}
	var existingKog *v1alpha2.Kog
	for _, kog := range kogList.Items {
		if kog.ObjectMeta.Name == kogName {
			existingKog = &kog
			break
		}
	}
	if existingKog == nil {
		return util.NewError("Could not find ioKog named " + kogName + " in namespace " + k8s.ns)
	}

	connectorExists := false
	for _, connector := range existingKog.Spec.Connectors.Instances {
		if connector.Name == name {
			connectorExists = true
			break
		}
	}
	if !connectorExists {
		existingKog.Spec.Connectors.Instances = append(existingKog.Spec.Connectors.Instances, v1alpha2.Connector{
			Name: name,
		})
	}
	existingKog.Spec.Connectors.Image = k8s.ms["connector"].containers[0].image

	err = cl.Update(context.Background(), existingKog)
	if err != nil {
		return err
	}

	return nil
}

const kogName = "iokog"

// CreateController on cluster
func (k8s *Kubernetes) CreateController(user IofogUser, replicas int) error {
	scheme := runtime.NewScheme()
	clientgoscheme.AddToScheme(scheme)
	crdapi.AddToScheme(scheme)

	cl, err := k8sclient.New(k8sconfig.GetConfigOrDie(), k8sclient.Options{Scheme: scheme})
	if err != nil {
		return err
	}

	// Check if kog exists
	kogKey := k8sclient.ObjectKey{
		Name:      kogName,
		Namespace: k8s.ns,
	}
	var kog v1alpha2.Kog
	found := true
	if err = cl.Get(context.Background(), kogKey, &kog); err != nil {
		if !k8serrors.IsNotFound(err) {
			return err
		}
		// Not found
		found = false
		kog = v1alpha2.Kog{
			ObjectMeta: metav1.ObjectMeta{
				Name:      kogName,
				Namespace: k8s.ns,
			},
		}
	}
	kog.Spec = v1alpha2.KogSpec{
		ControlPlane: v1alpha2.ControlPlane{
			IofogUser:              v1alpha2.IofogUser(user),
			ControllerReplicaCount: int32(replicas),
			ControllerImage:        k8s.ms["controller"].containers[0].image,
		},
		Connectors: v1alpha2.Connectors{
			Instances: []v1alpha2.Connector{},
		},
	}
	if found {
		if err = cl.Update(context.Background(), &kog); err != nil {
			return err
		}
	} else {
		if err = cl.Create(context.Background(), &kog); err != nil {
			return err
		}
	}

	return nil
}

func (k8s *Kubernetes) DeleteController() error {
	scheme := runtime.NewScheme()
	clientgoscheme.AddToScheme(scheme)
	crdapi.AddToScheme(scheme)

	cl, err := k8sclient.New(k8sconfig.GetConfigOrDie(), k8sclient.Options{Scheme: scheme})
	if err != nil {
		return err
	}

	kog := &v1alpha2.Kog{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kogName,
			Namespace: k8s.ns,
		},
	}
	if err = cl.Delete(context.Background(), kog); err != nil {
		return err
	}

	return nil
}

func (k8s *Kubernetes) DeleteConnector(name string) error {
	scheme := runtime.NewScheme()
	clientgoscheme.AddToScheme(scheme)
	crdapi.AddToScheme(scheme)

	cl, err := k8sclient.New(k8sconfig.GetConfigOrDie(), k8sclient.Options{Scheme: scheme})
	if err != nil {
		return err
	}

	kogList := &v1alpha2.KogList{}
	if err = cl.List(context.Background(), k8sclient.InNamespace(k8s.ns), kogList); err != nil {
		return err
	}
	if len(kogList.Items) == 0 {
		return util.NewError("Could not find existing ioKog on the Kubernetes cluster")
	}
	var existingKog *v1alpha2.Kog
	for _, kog := range kogList.Items {
		if kog.ObjectMeta.Name == kogName {
			existingKog = &kog
			break
		}
	}
	if existingKog == nil {
		return util.NewError("Could not find ioKog named " + kogName + " in namespace " + k8s.ns)
	}

	for idx, connector := range existingKog.Spec.Connectors.Instances {
		if connector.Name == name {
			instances := existingKog.Spec.Connectors.Instances
			existingKog.Spec.Connectors.Instances = append(instances[:idx], instances[idx+1:]...)
			if err = cl.Update(context.Background(), existingKog); err != nil {
				return err
			}
			break
		}
	}

	return nil
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

func (k8s *Kubernetes) SetControllerExternalDatabase(host, user, password, dbName string, port int) {
	k8s.ms["controller"].containers[0].env = []v1.EnvVar{
		{
			Name:  "DB_PROVIDER",
			Value: "postgres",
		},
		{
			Name:  "DB_NAME",
			Value: dbName,
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
	ip, err := k8s.waitForService("controller")
	if err != nil {
		return
	}
	endpoint = fmt.Sprintf("%s:%d", ip, k8s.ms["controller"].ports[0])
	return
}

func (k8s *Kubernetes) GetConnectorEndpoint(name string) (endpoint string, err error) {
	// TODO: This name formatting is magic that depends on the operator
	ip, err := k8s.waitForService("connector-" + name)
	if err != nil {
		return
	}
	endpoint = fmt.Sprintf("%s:%d", ip, k8s.ms["connector"].ports[0])
	return
}
