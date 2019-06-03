package iofog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/eclipse-iofog/cli/pkg/util"
	pb "github.com/schollz/progressbar"
	"k8s.io/api/core/v1"
	extsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"net/http"
	"os"
	"strings"
)

// Kubernetes struct to manage state of deployment on Kubernetes cluster
type Kubernetes struct {
	configFilename string
	clientset      *kubernetes.Clientset
	extsClientset  *extsclientset.Clientset
	crdName        string
	ns             string
}

// NewKubernetes constructs an object to manage cluster
func NewKubernetes(configFilename string) (*Kubernetes, error) {
	// Check if the kubeConfig file exists.
	_, err := os.Stat(configFilename)
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

	return &Kubernetes{
		configFilename: configFilename,
		clientset:      clientset,
		extsClientset:  extsClientset,
		crdName:        "iofogs.k8s.iofog.org",
		ns:             "iofog",
	}, nil
}

// CreateController on cluster
func (k8s *Kubernetes) CreateController() error {
	// Progress bar object
	pbCtx := progressBarContext{
		pb:    pb.New(100),
		quota: 90,
	}

	// Install ioFog Core
	token, ips, err := k8s.createCore(pbCtx)
	if err != nil {
		return err
	}

	pbCtx.quota = 10
	// Install ioFog K8s Extensions
	err = k8s.createExtension(token, ips, pbCtx)
	if err != nil {
		return err
	}

	return nil
}

// DeleteController from cluster
func (k8s *Kubernetes) DeleteController() error {
	// Progress bar object
	pb := pb.New(100)

	// Delete Deployments
	deps, err := k8s.clientset.AppsV1().Deployments(k8s.ns).List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, dep := range deps.Items {
		err = k8s.clientset.AppsV1().Deployments(k8s.ns).Delete(dep.Name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	pb.Add(10)

	// Delete Services
	svcs, err := k8s.clientset.CoreV1().Services(k8s.ns).List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, svc := range svcs.Items {
		err = k8s.clientset.CoreV1().Services(k8s.ns).Delete(svc.Name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	pb.Add(10)

	// Delete Service Accounts
	svcAccs, err := k8s.clientset.CoreV1().ServiceAccounts(k8s.ns).List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, acc := range svcAccs.Items {
		err = k8s.clientset.CoreV1().ServiceAccounts(k8s.ns).Delete(acc.Name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	pb.Add(10)

	// Delete Kubelet Cluster Role Binding
	err = k8s.clientset.RbacV1().ClusterRoleBindings().Delete(kubeletMicroservice.name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	pb.Add(10)

	// Delete Roles
	roles, err := k8s.clientset.RbacV1().Roles(k8s.ns).List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, role := range roles.Items {
		err = k8s.clientset.RbacV1().Roles(k8s.ns).Delete(role.Name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	pb.Add(10)

	// Delete Role Bindings
	roleBinds, err := k8s.clientset.RbacV1().RoleBindings(k8s.ns).List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, bind := range roleBinds.Items {
		err = k8s.clientset.RbacV1().RoleBindings(k8s.ns).Delete(bind.Name, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	pb.Add(10)

	// Delete CRD
	err = k8s.extsClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Delete(k8s.crdName, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	pb.Add(10)

	// Delete Namespace
	err = k8s.clientset.CoreV1().Namespaces().Delete(k8s.ns, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	pb.Add(30)

	return nil
}

func (k8s *Kubernetes) createCore(pbCtx progressBarContext) (token string, ips map[string]string, err error) {
	pbSlice := pbCtx.quota / 10

	// Create namespace
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: k8s.ns,
		},
	}
	_, _ = k8s.clientset.CoreV1().Namespaces().Create(ns)

	coreMs := []microservice{
		controllerMicroservice,
		connectorMicroservice,
	}
	// Create Controller and Connector Services and Pods
	for _, ms := range coreMs {
		svc := newService(k8s.ns, ms)
		_, err = k8s.clientset.CoreV1().Services(k8s.ns).Create(svc)
		if err != nil {
			return
		}
		svcAcc := newServiceAccount(k8s.ns, ms)
		_, err = k8s.clientset.CoreV1().ServiceAccounts(k8s.ns).Create(svcAcc)
		if err != nil {
			return
		}
		dep := newDeployment(k8s.ns, ms)
		_, err = k8s.clientset.AppsV1().Deployments(k8s.ns).Create(dep)
		if err != nil {
			return
		}
	}

	pbCtx.pb.Add(pbSlice)

	pbCtx.quota = pbSlice * 2
	// Wait for Controller and Connector Pods
	err = k8s.waitForPods(k8s.ns, pbCtx)
	if err != nil {
		return
	}

	pbCtx.quota = pbSlice * 4
	// Wait for Controller and Connector IPs and store them
	ips, err = k8s.waitForServices(k8s.ns, pbCtx)
	if err != nil {
		return
	}

	// Connect Controller to Connector
	podList, err := k8s.clientset.CoreV1().Pods(k8s.ns).List(metav1.ListOptions{LabelSelector: "name=controller"})
	if err != nil {
		return
	}
	podName := podList.Items[0].Name
	// TODO: (Serge) Get rid of this exec! Use REST API when implemented for this
	_, err = util.Exec("KUBECONFIG="+k8s.configFilename, "kubectl", "exec", podName, "-n", "iofog", "--", "node", "/controller/src/main", "connector", "add", "-n", "gke", "-d", "connector", "--dev-mode-on", "-i", ips["connector"])
	if err != nil {
		return
	}
	pbCtx.pb.Add(pbSlice)

	// Get Controller token through REST API
	contentType := "application/json"
	url := fmt.Sprintf("http://%s:%d/api/v3/", ips["controller"], controllerMicroservice.port)

	// TODO: (Serge) Create unique user?
	// Create user
	signupBody := strings.NewReader("{ \"firstName\": \"Dev\", \"lastName\": \"Test\", \"email\": \"user@domain.com\", \"password\": \"#Bugs4Fun\" }")
	resp, err := http.Post(url+"user/signup", contentType, signupBody)
	if err != nil {
		return
	}
	pbCtx.pb.Add(pbSlice)

	// Login user
	loginBody := strings.NewReader("{\"email\":\"user@domain.com\",\"password\":\"#Bugs4Fun\"}")
	resp, err = http.Post(url+"user/login", contentType, loginBody)
	if err != nil {
		return
	}

	// Read access token from HTTP response
	var auth map[string]interface{}
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	err = json.Unmarshal(buf.Bytes(), &auth)
	if err != nil {
		return
	}
	token, exists := auth["accessToken"].(string)
	if !exists {
		err = util.NewInternalError("Failed to get auth token from Controller")
		return
	}
	pbCtx.pb.Add(pbSlice)

	return
}

func (k8s *Kubernetes) createExtension(token string, ips map[string]string, pbCtx progressBarContext) error {
	pbSlice := pbCtx.quota / 5

	// Create Scheduler resources
	schedDep := newDeployment(k8s.ns, schedulerMicroservice)
	_, err := k8s.clientset.AppsV1().Deployments(k8s.ns).Create(schedDep)
	if err != nil {
		return err
	}
	schedAcc := newServiceAccount(k8s.ns, schedulerMicroservice)
	_, err = k8s.clientset.CoreV1().ServiceAccounts(k8s.ns).Create(schedAcc)
	if err != nil {
		return err
	}
	pbCtx.pb.Add(pbSlice)

	// Create Kubelet resources
	vkSvcAcc := newServiceAccount(k8s.ns, kubeletMicroservice)
	_, err = k8s.clientset.CoreV1().ServiceAccounts(k8s.ns).Create(vkSvcAcc)
	if err != nil {
		return err
	}
	pbCtx.pb.Add(pbSlice)

	vkRoleBind := newClusterRoleBinding(k8s.ns, kubeletMicroservice)
	_, err = k8s.clientset.RbacV1().ClusterRoleBindings().Create(vkRoleBind)
	if err != nil {
		return err
	}
	kubeletMicroservice.containers[0].args = []string{
		"--namespace",
		k8s.ns,
		"--iofog-token",
		token,
		"--iofog-url",
		fmt.Sprintf("http://%s:%d", ips["controller"], controllerMicroservice.port),
	}
	pbCtx.pb.Add(pbSlice)
	vkDep := newDeployment(k8s.ns, kubeletMicroservice)
	_, err = k8s.clientset.AppsV1().Deployments(k8s.ns).Create(vkDep)
	if err != nil {
		return err
	}

	// Create Operator resources
	opSvcAcc := newServiceAccount(k8s.ns, operatorMicroservice)
	_, err = k8s.clientset.CoreV1().ServiceAccounts(k8s.ns).Create(opSvcAcc)
	if err != nil {
		return err
	}
	opRole := newRole(k8s.ns, operatorMicroservice)
	_, err = k8s.clientset.RbacV1().Roles(k8s.ns).Create(opRole)
	if err != nil {
		return err
	}
	opRoleBind := newRoleBinding(k8s.ns, operatorMicroservice)
	_, err = k8s.clientset.RbacV1().RoleBindings(k8s.ns).Create(opRoleBind)
	if err != nil {
		return err
	}
	crd := newCustomResourceDefinition(k8s.crdName)
	_, err = k8s.extsClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
	if err != nil {
		return err
	}
	pbCtx.pb.Add(pbSlice)
	opDep := newDeployment(k8s.ns, operatorMicroservice)
	opDep.Spec.Template.Spec.Containers[0].Ports = []v1.ContainerPort{
		{
			ContainerPort: int32(operatorMicroservice.port),
			Name:          "metrics",
		},
	}
	opDep.Spec.Template.Spec.Containers[0].Command = []string{
		"iofog-operator",
	}
	opDep.Spec.Template.Spec.Containers[0].ReadinessProbe = &v1.Probe{
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
	}
	opDep.Spec.Template.Spec.Containers[0].Env = []v1.EnvVar{
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
			Value: operatorMicroservice.name,
		},
	}
	_, err = k8s.clientset.AppsV1().Deployments(k8s.ns).Create(opDep)
	if err != nil {
		return err
	}
	pbCtx.pb.Add(pbSlice)

	return nil
}

func (k8s *Kubernetes) waitForPods(namespace string, pbCtx progressBarContext) error {
	// Get Pods
	podList, err := k8s.clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	podCount := len(podList.Items)

	// Determine progress slice
	pbSlice := pbCtx.quota / podCount
	if pbSlice == 0 {
		pbSlice = 1
	}

	// Get watch handler to observe changes to pods
	watch, err := k8s.clientset.CoreV1().Pods(namespace).Watch(metav1.ListOptions{})
	if err != nil {
		return err
	}

	// Wait for cluster to be in ready state
	readyPods := make(map[string]bool, 0)
	// Keep reading events indefinitely
	for event := range watch.ResultChan() {
		// Get the pod
		pod, ok := event.Object.(*v1.Pod)
		if !ok {
			return util.NewInternalError("Failed to wait for pods in namespace: " + namespace)
		}
		// Check pod is in running state
		_, exists := readyPods[pod.Name]
		if !exists && pod.Status.Phase == "Running" {
			readyPods[pod.Name] = true
			pbCtx.pb.Add(pbSlice)
			// All pods are ready
			if len(readyPods) == podCount {
				watch.Stop()
			}
		}
	}
	return nil
}

func (k8s *Kubernetes) waitForServices(namespace string, pbCtx progressBarContext) (map[string]string, error) {
	// Get Services
	serviceList, err := k8s.clientset.CoreV1().Services(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	// Return ips of services upon completion
	serviceCount := len(serviceList.Items)
	ips := make(map[string]string, serviceCount)

	// Determine progress slice
	pbSlice := pbCtx.quota / serviceCount
	if pbSlice == 0 {
		pbSlice = 1
	}

	// Get watch handler to observe changes to services
	watch, err := k8s.clientset.CoreV1().Services(namespace).Watch(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// Wait for Services to have IPs allocated
	readyServices := make(map[string]bool, 0)
	for event := range watch.ResultChan() {
		svc, ok := event.Object.(*v1.Service)
		if !ok {
			return nil, util.NewInternalError("Failed to wait for services in namespace: " + namespace)
		}
		// Check if the Service has a LB with an IP
		_, exists := readyServices[svc.Name]
		ipCount := len(svc.Status.LoadBalancer.Ingress)
		if !exists && ipCount > 0 {
			// We don't expect multiple IPs for service, lets error here because could be undefined behaviour
			if ipCount != 1 {
				return nil, util.NewInternalError("Found unexpected number of IPs for service: " + svc.Name)
			}
			// Record the IP
			ips[svc.Name] = svc.Status.LoadBalancer.Ingress[0].IP
			readyServices[svc.Name] = true
			pbCtx.pb.Add(pbSlice)
			// All services are ready
			if len(readyServices) == serviceCount {
				watch.Stop()
			}
		}
	}

	return ips, nil
}

type progressBarContext struct {
	pb    *pb.ProgressBar
	quota int
}
