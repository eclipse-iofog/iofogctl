package iofog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/eclipse-iofog/cli/pkg/util"
	"k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/helm/pkg/helm"
	"net/http"
	"os"
	"strings"
)

// Kubernetes struct to manage state of deployment on Kubernetes cluster
type Kubernetes struct {
	configFilename string
	clientset      *kubernetes.Clientset
	helmClient     *helm.Client
	ns             string
	charts         [2]string
	chartVersion   string
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

	// Instantiate Helm client
	helmClient := helm.NewClient(helm.ConnectTimeout(15), helm.Host("35.197.185.110:44134"))

	return &Kubernetes{
		configFilename: configFilename,
		clientset:      clientset,
		helmClient:     helmClient,
		ns:             "iofog",
		charts:         [2]string{"iofog", "iofog-k8s"},
		chartVersion:   "0.1.0",
	}, nil
}

// Init is used to initialize the cluster
func (k8s *Kubernetes) Init() (err error) {
	// Ensure that pods in kube-system namespace are up and running
	err = k8s.waitForPods("kube-system")
	if err != nil {
		return
	}

	// Create namespace
	nsSpec := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: k8s.ns}}
	_, _ = k8s.clientset.CoreV1().Namespaces().Create(nsSpec)

	// Install Helm and associated resources
	err = k8s.initHelm()
	if err != nil {
		return
	}

	return nil
}

// CreateController on cluster
func (k8s *Kubernetes) CreateController() error {
	// Start Controller and Connector
	coreMs := []microservice{
		controllerMicroservice,
		connectorMicroservice,
	}
	for _, ms := range coreMs {
		svc := newService(k8s.ns, ms)
		_, err := k8s.clientset.CoreV1().Services(k8s.ns).Create(svc)
		if err != nil {
			return err
		}
		dep := newDeployment(k8s.ns, ms)
		_, err = k8s.clientset.AppsV1().Deployments(k8s.ns).Create(dep)
		if err != nil {
			return err
		}
	}

	// Wait for Controller and Connector Pods
	err := k8s.waitForPods(k8s.ns)
	if err != nil {
		return err
	}

	// Wait for Controller and Connector IPs and store them
	ips, err := k8s.waitForServices(k8s.ns)
	if err != nil {
		return err
	}

	// Connect Controller to Connector
	podList, err := k8s.clientset.CoreV1().Pods(k8s.ns).List(metav1.ListOptions{LabelSelector: "name=controller"})
	if err != nil {
		return err
	}
	podName := podList.Items[0].Name
	// TODO: (Serge) Get rid of this exec! Use REST API when implemented for this
	_, err = util.Exec("KUBECONFIG="+k8s.configFilename, "kubectl", "exec", podName, "-n", "iofog", "--", "node", "/controller/src/main", "connector", "add", "-n", "gke", "-d", "connector", "--dev-mode-on", "-i", ips["connector"])
	if err != nil {
		return err
	}

	// Get Controller token through REST API
	contentType := "application/json"
	url := fmt.Sprintf("http://%s:%d/api/v3/", ips["controller"], 51121)

	// TODO: (Serge) Create unique user?
	// Create user
	signupBody := strings.NewReader("{ \"firstName\": \"Dev\", \"lastName\": \"Test\", \"email\": \"user@domain.com\", \"password\": \"#Bugs4Fun\" }")
	resp, err := http.Post(url+"user/signup", contentType, signupBody)
	if err != nil {
		return err
	}

	// Login user
	loginBody := strings.NewReader("{\"email\":\"user@domain.com\",\"password\":\"#Bugs4Fun\"}")
	resp, err = http.Post(url+"user/login", contentType, loginBody)
	if err != nil {
		return err
	}

	// Read access token from HTTP response
	var auth map[string]interface{}
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	err = json.Unmarshal(buf.Bytes(), &auth)
	if err != nil {
		return err
	}
	token, exists := auth["accessToken"].(string)
	if !exists {
		return util.NewInternalError("Failed to get auth token from Controller")
	}

	// Install ioFog K8s Extensions
	_, err = util.Exec("KUBECONFIG="+k8s.configFilename, "helm", "install", "iofog/iofog-k8s", "--set-string", "controller.token="+token)
	if err != nil {
		return err
	}

	return nil
}

// DeleteController from cluster
func (k8s *Kubernetes) DeleteController() error {

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
	return nil
}

func (k8s *Kubernetes) waitForPods(namespace string) error {
	// Get Pods
	podList, err := k8s.clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	podCount := len(podList.Items)

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
			// All pods are ready
			if len(readyPods) == podCount {
				watch.Stop()
			}
		}
	}
	return nil
}

func (k8s *Kubernetes) waitForServices(namespace string) (map[string]string, error) {
	// Get Services
	serviceList, err := k8s.clientset.CoreV1().Services(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	// Return ips of services upon completion
	serviceCount := len(serviceList.Items)
	ips := make(map[string]string, serviceCount)

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
			// All services are ready
			if len(readyServices) == serviceCount {
				watch.Stop()
			}
		}
	}

	return ips, nil
}

func (k8s *Kubernetes) initHelm() error {
	// Check whether Helm already configured
	env := "KUBECONFIG=" + k8s.configFilename
	_, err := util.Exec(env, "helm", "list")
	if err == nil {
		return nil
	}

	// Create Tiller Service Account
	serviceAcc := &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: "tiller",
		},
	}
	_, err = k8s.clientset.CoreV1().ServiceAccounts("kube-system").Create(serviceAcc)
	if err != nil {
		return err
	}

	// Create Tiller Cluster Role Binding
	roleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "tiller-cluster-role",
		},
		RoleRef: rbacv1.RoleRef{Kind: "ClusterRole", Name: "cluster-admin"},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "tiller",
				Namespace: "kube-system",
			}},
	}
	_, err = k8s.clientset.RbacV1().ClusterRoleBindings().Create(roleBinding)
	if err != nil {
		return err
	}

	// Execute Helm init commands
	_, err = util.Exec(env, "helm", "init", "--wait", "--service-account", "tiller")
	if err != nil {
		return err
	}
	_, err = util.Exec(env, "helm", "repo", "add", "iofog", "https://eclipse-iofog.github.io/helm")
	if err != nil {
		return err
	}
	_, err = util.Exec(env, "helm", "repo", "update", "iofog")
	if err != nil {
		return err
	}

	return nil
}
