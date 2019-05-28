package iofog

import ( 
	"github.com/eclipse-iofog/cli/pkg/util"
	"os"
	//"k8s.io/helm/pkg/helm"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

// Kubernetes struct to manage state of deployment on Kubernetes cluster
type Kubernetes struct {
	clientset  *kubernetes.Clientset
	ns string
}

func NewKubernetes(configFilename string) *Kubernetes {

	// Check if the kubeConfig file exists.
	_, err := os.Stat(configFilename)
	util.Check(err)

	// Get the kubernetes config from the filepath.
	config, err := clientcmd.BuildConfigFromFlags("", configFilename)
	util.Check(err)

	clientset, err := kubernetes.NewForConfig(config)
	util.Check(err)

	return &Kubernetes{
		clientset: clientset,
		ns: "iofog",
	}
}

func (k8s *Kubernetes) Init() (err error) {
	// Get kube-system pod details
	podList, err := k8s.clientset.CoreV1().Pods("kube-system").List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	watch, err := k8s.clientset.CoreV1().Pods("kube-system").Watch(metav1.ListOptions{})
	if err != nil {
		return
	}

	// Wait for cluster to be in ready state
	readyPods := make(map[string]bool, 0)
	for event := range watch.ResultChan() {
        p, ok := event.Object.(*v1.Pod)
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