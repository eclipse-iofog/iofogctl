package iofog

import (
	"github.com/eclipse-iofog/cli/pkg/util"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/helm/pkg/helm"
	"os"
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

func (k8s *Kubernetes) Init() (err error) {
	err = k8s.waitForPods("kube-system")
	if err != nil {
		return
	}

	// Create namespace
	nsSpec := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: k8s.ns}}
	_, err = k8s.clientset.CoreV1().Namespaces().Create(nsSpec)
	if err != nil {
		return
	}

	err = k8s.initHelm()
	if err != nil {
		return
	}

	return nil
}

func (k8s *Kubernetes) Clean() error {
	return nil
}

// Deploy Controller to Kubernetes cluster
func (k8s *Kubernetes) CreateController() error {
	return nil
}

func (k8s *Kubernetes) DeleteController() error {
	return nil
}

func (k8s *Kubernetes) waitForPods(namespace string) error {
	// Get kube-system pod details
	podList, err := k8s.clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	watch, err := k8s.clientset.CoreV1().Pods(namespace).Watch(metav1.ListOptions{})
	if err != nil {
		return err
	}

	// Wait for cluster to be in ready state
	readyPods := make(map[string]bool, 0)
	for event := range watch.ResultChan() {
		p, ok := event.Object.(*corev1.Pod)
		if !ok {
			return util.NewInternalError("Failed to wait for Kubernetes cluster to be ready")
		}
		_, exists := readyPods[p.Name]
		if !exists && p.Status.Phase == "Running" {
			readyPods[p.Name] = true
			if len(readyPods) == len(podList.Items) {
				watch.Stop()
			}
		}
	}
	return nil
}

func (k8s *Kubernetes) initHelm() error {
	// Create Tiller Service Account
	serviceAcc := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: "tiller",
		},
	}
	_, err := k8s.clientset.CoreV1().ServiceAccounts("kube-system").Create(serviceAcc)
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
	env := "KUBECONFIG=" + k8s.configFilename
	err = util.Exec(env, "helm", "init", "--wait", "--service-account", "tiller")
	if err != nil {
		return err
	}
	err = util.Exec(env, "helm", "repo", "add", "iofog", "https://eclipse-iofog.github.io/helm")
	if err != nil {
		return err
	}
	err = util.Exec(env, "helm", "repo", "update", "iofog")
	if err != nil {
		return err
	}

	return nil
}
