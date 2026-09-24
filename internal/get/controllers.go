package get

import (
	// "fmt"
	"time"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type controllerExecutor struct {
	namespace string
}

func newControllerExecutor(namespace string) *controllerExecutor {
	c := &controllerExecutor{}
	c.namespace = namespace
	return c
}

func (exe *controllerExecutor) GetName() string {
	return ""
}

func (exe *controllerExecutor) Execute() error {
	table, err := generateControllerOutput(exe.namespace)
	if err != nil {
		return err
	}
	printNamespace(exe.namespace)
	return print(table)
}

func generateControllerOutput(namespace string) (table [][]string, err error) {
	// Get controller config details
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return
	}

	podStatuses := []string{}
	// Handle k8s
	baseControlPlane, err := ns.GetControlPlane()
	if err != nil {
		if rsc.IsNoControlPlaneError(err) {
			err = nil
		} else {
			return
		}
	}
	if controlPlane, ok := baseControlPlane.(*rsc.KubernetesControlPlane); ok {
		if err = updateControllerPods(controlPlane, namespace); err != nil {
			return
		}
		ns.SetControlPlane(controlPlane)
		if err = config.Flush(); err != nil {
			return
		}
		for idx := range controlPlane.ControllerPods {
			podStatuses = append(podStatuses, controlPlane.ControllerPods[idx].Status)
		}
	}

	// Handle remote and local
	controllers := ns.GetControllers()

	// Generate table and headers
	table = make([][]string, len(controllers)+1)
	headers := []string{"CONTROLLER", "STATUS", "AGE", "UPTIME", "VERSION", "ADDR", "PORT"}
	table[0] = append(table[0], headers...)

	if len(controllers) == 0 {
		return table, nil
	}

	uptime := "-"
	version := "-"
	apiStatus := "Failing"
	ctrl, err := clientutil.NewControllerClient(namespace)
	if err != nil {
		return table, err
	}
	ctrlStatus, err := ctrl.GetStatus()
	if err == nil {
		uptime = util.FormatDuration(time.Duration(int64(ctrlStatus.UptimeSeconds)) * time.Second)
		apiStatus = ctrlStatus.Status
		version = ctrlStatus.Versions.Controller
	}

	// Populate rows — reuse ADDR/PORT from config and VERSION/UPTIME from a single GetStatus()
	for idx, ctrlConfig := range controllers {
		status := apiStatus
		if idx < len(podStatuses) {
			status = podStatuses[idx]
		}

		age := "-"
		if ctrlConfig.GetCreatedTime() != "" {
			age, _ = util.ElapsedUTC(ctrlConfig.GetCreatedTime(), util.NowUTC())
		}

		addr, port := getAddressAndPort(ctrlConfig.GetEndpoint(), client.ControllerPortString)
		row := []string{
			ctrlConfig.GetName(),
			status,
			age,
			uptime,
			version,
			addr,
			port,
		}
		table[idx+1] = append(table[idx+1], row...)
	}

	return table, nil
}

func updateControllerPods(controlPlane *rsc.KubernetesControlPlane, namespace string) (err error) {
	// Clear existing
	controlPlane.ControllerPods = []rsc.KubernetesController{}
	// Get pods
	installer, err := install.NewKubernetes(controlPlane.KubeConfig, namespace)
	if err != nil {
		return
	}

	// Set HTTPS configuration if present in the control plane
	if controlPlane.Controller.Https != nil {
		installer.SetHttpsEnabled(controlPlane.Controller.Https)
	}

	pods, err := installer.GetControllerPods()
	if err != nil {
		return
	}
	// Add pods
	for idx := range pods {
		k8sPod := rsc.KubernetesController{
			Endpoint: controlPlane.Endpoint,
			PodName:  pods[idx].Name,
			Status:   pods[idx].Status,
		}
		if err := controlPlane.AddController(&k8sPod); err != nil {
			return err
		}
	}
	return
}
