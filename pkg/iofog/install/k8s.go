package install

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"reflect"
	"strings"
	"time"

	ioclient "github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	iofogv3 "github.com/eclipse-iofog/iofog-operator/v3/apis"
	cpv3 "github.com/eclipse-iofog/iofog-operator/v3/apis/controlplanes/v3"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	extsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	extsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // GCP auth
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	opclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ECNname
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
	services      cpv3.Services
	images        cpv3.Images
	ingresses     cpv3.Ingresses
	httpsEnabled  *bool // Store HTTPS configuration
	isViewerDns   *bool // Store isViewerDns configuration
	// router        cpv3.Router
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

func (k8s *Kubernetes) SetOperatorImage(image string) {
	if image != "" {
		k8s.operator.containers[0].image = image
	} else {
		k8s.operator.containers[0].image = util.GetOperatorImage()
	}
}

func (k8s *Kubernetes) SetRouterImage(image string) {
	if image != "" {
		k8s.images.Router = image
	} else {
		k8s.images.Router = util.GetRouterImage()
	}
}

func (k8s *Kubernetes) SetControllerImage(image string) {
	if image != "" {
		k8s.images.Controller = image
	} else {
		k8s.images.Controller = util.GetControllerImage()
	}
}

func (k8s *Kubernetes) SetNatsImage(image string) {
	if image != "" {
		k8s.images.Nats = image
	} else {
		k8s.images.Nats = util.GetNatsImage()
	}
}

func (k8s *Kubernetes) SetPullSecret(pullSecret string) {
	if pullSecret != "" {
		k8s.images.PullSecret = pullSecret
		k8s.operator.imagePullSecret = pullSecret
	}
}

func (k8s *Kubernetes) SetHttpsEnabled(enabled *bool) {
	k8s.httpsEnabled = enabled
}

func (k8s *Kubernetes) SetIsViewerDns(enabled *bool) {
	k8s.isViewerDns = enabled
}

func (k8s *Kubernetes) enableCustomResources() error {
	ctx := context.Background()
	// Control Plane and App
	for _, crd := range []*extsv1.CustomResourceDefinition{iofogv3.NewControlPlaneCustomResource(), iofogv3.NewAppCustomResource()} {
		// Try create new
		if _, err := k8s.extsClientset.ApiextensionsV1().CustomResourceDefinitions().Create(ctx, crd, metav1.CreateOptions{}); err != nil {
			if !k8serrors.IsAlreadyExists(err) {
				return err
			}
			// Update
			existingCRD, err := k8s.extsClientset.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, crd.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}

			// Always update the CRD if:
			// 1. The CRD is not supported (major version mismatch)
			// 2. The versions array is different (new features/types added)
			shouldUpdate := !iofogv3.IsSupportedCustomResource(existingCRD) ||
				!reflect.DeepEqual(existingCRD.Spec.Versions, crd.Spec.Versions)

			if shouldUpdate {
				// Preserve the existing status
				existingCRD.Spec.Versions = crd.Spec.Versions

				// Update the CRD
				if _, err := k8s.extsClientset.ApiextensionsV1().CustomResourceDefinitions().Update(ctx, existingCRD, metav1.UpdateOptions{}); err != nil {
					return err
				}
			}
		}
	}

	// Deploy operator again
	if err := k8s.createOperator(); err != nil {
		return err
	}

	// Enable client
	if err := k8s.enableOperatorClient(); err != nil {
		return err
	}

	return nil
}

func (k8s *Kubernetes) enableOperatorClient() (err error) {
	scheme := iofogv3.InitClientScheme()
	k8s.opClient, err = opclient.New(k8s.config, opclient.Options{Scheme: scheme})
	if err != nil {
		return err
	}
	return nil
}

// CreateController on cluster
func (k8s *Kubernetes) CreateControlPlane(conf *K8SControllerConfig) (endpoint string, err error) {
	// Create namespace if required
	Verbose("Creating namespace " + k8s.ns)
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: k8s.ns,
		},
	}
	if _, err = k8s.clientset.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{}); err != nil {
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
	var cp cpv3.ControlPlane
	found := true
	if err = k8s.opClient.Get(context.Background(), cpKey, &cp); err != nil {
		if !k8serrors.IsNotFound(err) {
			return
		}
		// Not found, set basic info
		found = false
		cp = cpv3.ControlPlane{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cpInstanceName,
				Namespace: k8s.ns,
			},
		}
	}

	// Set specification
	cp.Spec.Replicas.Controller = conf.Replicas
	if conf.ReplicasNats >= 2 {
		cp.Spec.Replicas.Nats = conf.ReplicasNats
	}
	cp.Spec.Database = cpv3.Database(conf.Database)
	cp.Spec.Auth = cpv3.Auth(conf.Auth)
	cp.Spec.Events = cpv3.Events(conf.Events)
	// cp.Spec.User = cpv3.User(conf.User)
	cp.Spec.Services = k8s.services
	cp.Spec.Ingresses = k8s.ingresses
	cp.Spec.Images = k8s.images
	if conf.Nats != nil {
		cp.Spec.Nats = conf.Nats
	}
	if conf.Vault != nil {
		cp.Spec.Vault = conf.Vault
	}
	// cp.Spec.Router = k8s.router
	cp.Spec.Controller.EcnViewerPort = conf.EcnViewerPort
	cp.Spec.Controller.EcnViewerURL = conf.EcnViewerURL
	cp.Spec.Controller.LogLevel = conf.LogLevel
	cp.Spec.Controller.PidBaseDir = conf.PidBaseDir
	cp.Spec.Controller.Https = conf.Https
	cp.Spec.Controller.SecretName = conf.SecretName

	// Store HTTPS configuration for endpoint generation
	k8s.SetHttpsEnabled(conf.Https)

	// Create or update Control Plane
	if found {
		Verbose("Updating existing Control Plane")
		if err = k8s.opClient.Update(context.Background(), &cp); err != nil {
			return
		}
	} else {
		cp.SetConditionDeploying(nil)
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
	case <-time.After(600 * time.Second):
		err = util.NewInternalError("Failed to wait for Default Router registration")
	}

	return endpoint, err
}

func (k8s *Kubernetes) getReadyPod() (readyPod *corev1.Pod, err error) {
	// Check operator logs
	pods, err := k8s.clientset.CoreV1().Pods(k8s.ns).List(context.Background(), metav1.ListOptions{
		LabelSelector: "name=iofog-operator", // TODO: Decouple this
	})
	if err != nil {
		return
	}
	if len(pods.Items) == 0 {
		err = util.NewInternalError("Could not find any Operator Pods ")
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
	return pod, err
}

// Watch Operator logs
// Report error from Operator if found in logs
// Operator Pods are deleted and created when Control Plane redeployed
func (k8s *Kubernetes) monitorOperator(errCh chan error) {
	errSuffix := "while awaiting finalization of Control Plane"
	for {
		time.Sleep(2 * time.Second)
		pod, err := k8s.getReadyPod()
		if err != nil {
			errCh <- fmt.Errorf("%s %s", err.Error(), errSuffix)
			return
		}
		// Could not find ready Operator Pod
		if pod == nil {
			continue
		}
		// Get the logs of ready Pod
		req := k8s.clientset.CoreV1().Pods(k8s.ns).GetLogs(pod.Name, &corev1.PodLogOptions{})
		podLogs, err := req.Stream(context.Background())
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

		// Check controlplane resource status
		var cp cpv3.ControlPlane
		if err = k8s.opClient.Get(context.Background(), opclient.ObjectKey{
			Name:      cpInstanceName,
			Namespace: k8s.ns,
		}, &cp); err != nil {
			errCh <- util.NewInternalError("Error reading Control Plane resource " + errSuffix)
			return
		}

		if cp.IsReady() {
			errCh <- nil
			return
		}

		// podLogsStr := buf.String()
		// Allow errors, need to fix iofog-operator to not have errors anymore
		// errDelim := `ERROR` // TODO: Decouple iofogctl-operator err string
		// if strings.Contains(podLogsStr, errDelim) {
		// 	msg := ""
		// 	logLines := strings.Split(podLogsStr, "\n")
		// 	for _, line := range logLines {
		// 		// Error line pertains to this NS?
		// 		if strings.Contains(line, errDelim) && strings.Contains(line, fmt.Sprintf(`"namespace": "%s"`, k8s.ns)) {
		// 			msg = fmt.Sprintf("%s\n%s", msg, line)
		// 			errCh <- util.NewInternalError("Operator failed to reconcile Control Plane " + msg)
		// 			return
		// 		}
		// 	}
		// Error pertains to another Control Plane
		// }

		// Continue loop, wait for Router registration or error...
	}
}

func (k8s *Kubernetes) deleteOperator() (err error) {
	// Resource name for deletions
	name := k8s.operator.name

	ctx := context.Background()

	// Service Account
	if err = k8s.clientset.CoreV1().ServiceAccounts(k8s.ns).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		if !k8serrors.IsNotFound(err) {
			return
		}
	}

	// Role
	if err = k8s.clientset.RbacV1().Roles(k8s.ns).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		if !k8serrors.IsNotFound(err) {
			return
		}
	}

	// Role Binding
	if err = k8s.clientset.RbacV1().RoleBindings(k8s.ns).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		if !k8serrors.IsNotFound(err) {
			return
		}
	}

	// Deployment
	if err = k8s.clientset.AppsV1().Deployments(k8s.ns).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		if !k8serrors.IsNotFound(err) {
			return
		}
	}

	return nil
}

func (k8s *Kubernetes) createOperator() (err error) {
	ctx := context.Background()

	// Service Account
	opSvcAcc := newServiceAccount(k8s.ns, k8s.operator)
	if _, err = k8s.clientset.CoreV1().ServiceAccounts(k8s.ns).Create(ctx, opSvcAcc, metav1.CreateOptions{}); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return
		}
	}

	// Role
	role := newRole(k8s.ns, k8s.operator)
	if _, err = k8s.clientset.RbacV1().Roles(k8s.ns).Create(ctx, role, metav1.CreateOptions{}); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return
		}
	}

	// Role Binding
	rb := newRoleBinding(k8s.ns, k8s.operator)
	if _, err = k8s.clientset.RbacV1().RoleBindings(k8s.ns).Create(ctx, rb, metav1.CreateOptions{}); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return
		}
	}

	// Deployment
	opDep := newDeployment(k8s.ns, k8s.operator)
	if _, err = k8s.clientset.AppsV1().Deployments(k8s.ns).Create(ctx, opDep, metav1.CreateOptions{}); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return
		}
		// Redeploy the operator
		if err = k8s.clientset.AppsV1().Deployments(k8s.ns).Delete(ctx, k8s.operator.name, metav1.DeleteOptions{}); err != nil {
			return
		}
		if _, err = k8s.clientset.AppsV1().Deployments(k8s.ns).Create(ctx, opDep, metav1.CreateOptions{}); err != nil {
			return
		}
	}
	return nil
}

func (k8s *Kubernetes) DeleteControlPlane() error {
	// Prepare Control Plane client
	if err := k8s.enableOperatorClient(); err != nil {
		return err
	}

	// Delete Control Plane
	cp := &cpv3.ControlPlane{
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
		if err := k8s.clientset.CoreV1().Namespaces().Delete(context.Background(), k8s.ns, metav1.DeleteOptions{}); err != nil {
			if !k8serrors.IsNotFound(err) {
				return err
			}
		}
	}

	return nil
}

func (k8s *Kubernetes) waitForService(name string, targetPort int32) (addr string, nodePort int32, err error) {
	// Get watch handler to observe changes to services
	watch, err := k8s.clientset.CoreV1().Services(k8s.ns).Watch(context.Background(), metav1.ListOptions{})
	if err != nil {
		return
	}
	defer watch.Stop()

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
			addr, nodePort = k8s.handleLoadBalancer(svc, targetPort)
			if addr == "" {
				continue
			}
			return

		case corev1.ServiceTypeNodePort:
			addr, err = k8s.getNodePortAddress(name)
			if err != nil {
				util.PrintNotify("Could not get an external IP address of any Kubernetes nodes for NodePort service " + name + "\nTrying to reach the cluster IP of the service")
				addr, err = k8s.getClusterIPAddress(name)
				if err != nil {
					return
				}
			}
			nodePort, err = k8s.getPort(svc, name, targetPort)
			return
		case corev1.ServiceTypeClusterIP:
			// Ingress must be ready for ClusterIP service type
			addr, err = k8s.waitForIngress("iofog-controller")
			if err != nil {
				util.PrintNotify("Failed to handle Ingress for ClusterIP service")
				continue
			}
			if addr == "" {
				continue
			}
			nodePort = targetPort
			return
		default:
			err = util.NewError("Found Service was not of supported type")
			return
		}
	}
	err = util.NewError("Did not receive any events from Kuberenetes API Server")
	return addr, nodePort, err
}

func (k8s *Kubernetes) waitForIngress(name string) (addr string, err error) {
	// Create a watch to observe changes to Ingress resources
	watcher, err := k8s.clientset.NetworkingV1().Ingresses(k8s.ns).Watch(context.Background(), metav1.ListOptions{})
	if err != nil {
		err = util.NewError("Failed to create watch for Ingress: " + err.Error())
		return
	}
	defer watcher.Stop()

	// Process events from the watch
	for event := range watcher.ResultChan() {
		ingress, ok := event.Object.(*networkingv1.Ingress)
		if !ok {
			err = util.NewInternalError("Failed to parse Ingress event")
			return
		}

		// Check if the Ingress resource matches the name we're waiting for
		if ingress.Name == name {
			// Check if Ingress has rules
			if len(ingress.Spec.Rules) > 0 {
				host := ingress.Spec.Rules[0].Host
				addr = "https://" + host
				return
			}

			// Ingress found but has no rules, continue waiting
			util.PrintNotify("Ingress resource found but no rules present, continuing to watch...")
		}
	}

	err = util.NewError("Did not receive any valid Ingress events")
	return
}

func (k8s *Kubernetes) getPort(svc *corev1.Service, name string, targetPort int32) (nodePort int32, err error) {
	// Get the port allocated on the node
	for _, port := range svc.Spec.Ports {
		if port.TargetPort.IntVal == targetPort {
			nodePort = port.NodePort
			break
		}
	}
	if nodePort == 0 {
		err = util.NewError("Could not get node port for Kubernetes service " + name)
		return
	}
	return
}

func (k8s *Kubernetes) getClusterIPAddress(name string) (addr string, err error) {
	// Get a list of K8s nodes and return one of their external IPs
	var nodeList *corev1.NodeList
	nodeList, err = k8s.clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return
	}
	for idx := range nodeList.Items {
		node := &nodeList.Items[idx]
		for _, addrs := range node.Status.Addresses {
			if addrs.Type == corev1.NodeInternalIP {
				addr = addrs.Address
				break
			}
		}
	}
	if addr == "" {
		err = util.NewError("Could not get address for ClusterIP " + name)
	}
	return
}

func (k8s *Kubernetes) getNodePortAddress(name string) (addr string, err error) {
	// Get a list of K8s nodes and return one of their external IPs
	var nodeList *corev1.NodeList
	nodeList, err = k8s.clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return
	}
	if len(nodeList.Items) == 0 {
		err = util.NewError("Could not find Kubernetes nodes when waiting for NodePort service " + name)
		return
	}
	// Return external IP of any of the nodes in the cluster
	for idx := range nodeList.Items {
		node := &nodeList.Items[idx]
		for _, addrs := range node.Status.Addresses {
			if addrs.Type == corev1.NodeExternalIP {
				addr = addrs.Address
				break
			}
		}
	}
	if addr == "" {
		err = util.NewError("Could not find address in Node Port service " + name)
	}
	return
}

func (k8s *Kubernetes) handleLoadBalancer(svc *corev1.Service, targetPort int32) (addr string, nodePort int32) {
	nodePort = targetPort
	// TODO: error if Ingress len == 0
	ip := svc.Status.LoadBalancer.Ingress[0].IP
	host := svc.Status.LoadBalancer.Ingress[0].Hostname
	if ip != "" {
		addr = ip
	}
	if host != "" {
		addr = host
	}
	return
}

func (k8s *Kubernetes) SetControllerService(svcType, address string, annotations map[string]string, externalTrafficPolicy string) {
	if svcType != "" {
		k8s.services.Controller.Type = svcType
	} else {
		k8s.services.Controller.Type = string(corev1.ServiceTypeLoadBalancer)
	}
	k8s.services.Controller.Address = address
	k8s.services.Controller.Annotations = annotations
	k8s.services.Controller.ExternalTrafficPolicy = externalTrafficPolicy
}

func (k8s *Kubernetes) SetRouterService(svcType, address string, annotations map[string]string, externalTrafficPolicy string) {
	if svcType != "" {
		k8s.services.Router.Type = svcType
	} else {
		k8s.services.Router.Type = string(corev1.ServiceTypeLoadBalancer)
	}
	k8s.services.Router.Address = address
	k8s.services.Router.Annotations = annotations
	k8s.services.Router.ExternalTrafficPolicy = externalTrafficPolicy
}

func (k8s *Kubernetes) SetControllerIngress(annotations map[string]string, ingressClassName string, host string, secretName string) {
	k8s.ingresses.Controller.Annotations = annotations
	k8s.ingresses.Controller.IngressClassName = ingressClassName
	k8s.ingresses.Controller.Host = host
	k8s.ingresses.Controller.SecretName = secretName
}

func (k8s *Kubernetes) SetRouterIngress(address string, messagePort int, interiorPort int, edgePort int) {
	k8s.ingresses.Router.Address = address
	k8s.ingresses.Router.MessagePort = messagePort
	k8s.ingresses.Router.InteriorPort = interiorPort
	k8s.ingresses.Router.EdgePort = edgePort
}

func (k8s *Kubernetes) SetNatsService(svcType, address string, annotations map[string]string, externalTrafficPolicy string) {
	if svcType != "" {
		k8s.services.Nats.Type = svcType
	} else {
		k8s.services.Nats.Type = string(corev1.ServiceTypeClusterIP)
	}
	k8s.services.Nats.Address = address
	k8s.services.Nats.Annotations = annotations
	k8s.services.Nats.ExternalTrafficPolicy = externalTrafficPolicy
}

func (k8s *Kubernetes) SetNatsServerService(svcType, address string, annotations map[string]string, externalTrafficPolicy string) {
	if svcType != "" {
		k8s.services.NatsServer.Type = svcType
	} else {
		k8s.services.NatsServer.Type = string(corev1.ServiceTypeLoadBalancer)
	}
	k8s.services.NatsServer.Address = address
	k8s.services.NatsServer.Annotations = annotations
	k8s.services.NatsServer.ExternalTrafficPolicy = externalTrafficPolicy
}

func (k8s *Kubernetes) SetNatsIngress(address string, serverPort, clusterPort, leafPort, mqttPort, httpPort int) {
	k8s.ingresses.Nats.Address = address
	k8s.ingresses.Nats.ServerPort = serverPort
	k8s.ingresses.Nats.ClusterPort = clusterPort
	k8s.ingresses.Nats.LeafPort = leafPort
	k8s.ingresses.Nats.MqttPort = mqttPort
	k8s.ingresses.Nats.HttpPort = httpPort
}

// func (k8s *Kubernetes) SetRouterConfig(HA *bool) {
// 	k8s.router.HA = HA

// }

func (k8s *Kubernetes) ExistsInNamespace(namespace string) error {
	ctx := context.Background()
	// Check namespace exists
	if _, err := k8s.clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			return util.NewError("Could not find Namespace " + namespace + " on Kubernetes cluster")
		}
		return err
	}

	// Check services exist
	svcList, err := k8s.clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for idx := range svcList.Items {
		svc := &svcList.Items[idx]
		if svc.Name == controller {
			return nil
		}
	}
	return util.NewError("Could not find Controller Service in Kubernetes namespace " + namespace)
}

func (k8s *Kubernetes) formatEndpoint(endpoint string, port int32, isViewerDns ...bool) (*url.URL, error) {
	// Ensure protocol
	if !strings.Contains(endpoint, "://") {
		// Check if HTTPS should be used
		if k8s.httpsEnabled != nil && *k8s.httpsEnabled {
			endpoint = fmt.Sprintf("https://%s", endpoint)
		} else {
			endpoint = fmt.Sprintf("http://%s", endpoint)
		}
	}
	URL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	// Ensure port is added if not present
	// if !strings.Contains(URL.Host, ":") {
	// Ensure port when scheme is not HTTPS OR when isViewerDns is false
	if !strings.Contains(URL.Host, ":") && ((URL.Scheme != "https") || (len(isViewerDns) > 0 && !isViewerDns[0])) {

		URL.Host += fmt.Sprintf(":%d", port)
	}
	return URL, nil

}

func (k8s *Kubernetes) GetControllerEndpoint() (endpoint string, err error) {
	ip, port, err := k8s.waitForService(controller, ioclient.ControllerPort)
	if err != nil {
		return "", err
	}
	isViewerDns := false
	if k8s.isViewerDns != nil && *k8s.isViewerDns {
		isViewerDns = true
	}
	formattedURL, err := k8s.formatEndpoint(ip, port, isViewerDns)
	if err != nil {
		return "", err
	}
	endpoint = formattedURL.String()

	// Check if HTTPS is enabled
	useHTTPS := false
	if k8s.httpsEnabled != nil && *k8s.httpsEnabled {
		useHTTPS = true
	}

	return util.GetControllerEndpoint(endpoint, useHTTPS)
}

func (k8s *Kubernetes) GetControllerPods() (podNames []Pod, err error) {
	podNames = []Pod{}
	// List pods
	pods, err := k8s.clientset.CoreV1().Pods(k8s.ns).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return
	}
	// Find Controller pods
	for idx := range pods.Items {
		if pods.Items[idx].Labels["iofog.org/component"] == controller {
			podNames = append(podNames, Pod{
				Name:   pods.Items[idx].Name,
				Status: string(pods.Items[idx].Status.Phase),
			})
		}
	}
	return
}
