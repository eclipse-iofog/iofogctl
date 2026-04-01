package deployairgap

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

// RequiredImages represents all images needed for airgap deployment
type RequiredImages struct {
	Controller string
	Agent      string
	RouterX86  string
	RouterARM  string
	Nats       string
	Debugger   string
}

// getCatalogItemByName tries the given name, then fallbackNames if the first lookup fails (e.g. casing).
func getCatalogItemByName(clt *client.Client, name string, fallbackNames ...string) (item *client.CatalogItemInfo, err error) {
	item, err = clt.GetCatalogItemByName(name)
	if err == nil {
		return item, nil
	}
	for _, n := range fallbackNames {
		item, err = clt.GetCatalogItemByName(n)
		if err == nil {
			return item, nil
		}
	}
	return nil, err
}

// applyRouterImagesFromCatalog sets RouterX86 and RouterARM from catalog item images.
func applyRouterImagesFromCatalog(images *RequiredImages, item *client.CatalogItemInfo) {
	if item == nil || len(item.Images) == 0 {
		return
	}
	for _, img := range item.Images {
		switch client.AgentTypeIDAgentTypeDict[img.AgentTypeID] {
		case "x86":
			images.RouterX86 = img.ContainerImage
		case "arm":
			images.RouterARM = img.ContainerImage
		}
	}
}

// applyDebuggerImageFromCatalog sets Debugger from catalog item (first x86 or arm image).
func applyDebuggerImageFromCatalog(images *RequiredImages, item *client.CatalogItemInfo) {
	if item == nil || len(item.Images) == 0 {
		return
	}
	for _, img := range item.Images {
		if client.AgentTypeIDAgentTypeDict[img.AgentTypeID] == "x86" || client.AgentTypeIDAgentTypeDict[img.AgentTypeID] == "arm" {
			images.Debugger = img.ContainerImage
			return
		}
	}
}

// applyNatsImageFromCatalog sets Nats from catalog item (first image).
func applyNatsImageFromCatalog(images *RequiredImages, item *client.CatalogItemInfo) {
	if item == nil || len(item.Images) == 0 {
		return
	}
	images.Nats = item.Images[0].ContainerImage
}

// applyYAMLFallbackForController fills any empty router/nats/debugger from controlPlane.SystemMicroservices; util is last fallback.
func applyYAMLAndUtilFallbackForController(images *RequiredImages, controlPlane *rsc.RemoteControlPlane) {
	if controlPlane == nil {
		return
	}
	sm := &controlPlane.SystemMicroservices
	if images.RouterX86 == "" {
		if sm.Router.X86 != "" {
			images.RouterX86 = sm.Router.X86
		} else {
			images.RouterX86 = util.GetRouterImage()
		}
	}
	if images.RouterARM == "" {
		if sm.Router.ARM != "" {
			images.RouterARM = sm.Router.ARM
		} else {
			images.RouterARM = util.GetRouterARMImage()
		}
	}
	if images.Nats == "" {
		if sm.Nats.X86 != "" {
			images.Nats = sm.Nats.X86
		} else if sm.Nats.ARM != "" {
			images.Nats = sm.Nats.ARM
		} else {
			images.Nats = util.GetNatsImage()
		}
	}
	if images.Debugger == "" {
		images.Debugger = util.GetDebuggerImage()
	}
}

// applyYAMLAndUtilFallbackForAgent fills any empty router/nats/debugger from controlPlane (if non-nil) then util.
func applyYAMLAndUtilFallbackForAgent(images *RequiredImages, controlPlane *rsc.RemoteControlPlane) {
	if images.RouterX86 == "" {
		if controlPlane != nil && controlPlane.SystemMicroservices.Router.X86 != "" {
			images.RouterX86 = controlPlane.SystemMicroservices.Router.X86
		} else {
			images.RouterX86 = util.GetRouterImage()
		}
	}
	if images.RouterARM == "" {
		if controlPlane != nil && controlPlane.SystemMicroservices.Router.ARM != "" {
			images.RouterARM = controlPlane.SystemMicroservices.Router.ARM
		} else {
			images.RouterARM = util.GetRouterARMImage()
		}
	}
	if images.Nats == "" {
		if controlPlane != nil {
			if controlPlane.SystemMicroservices.Nats.X86 != "" {
				images.Nats = controlPlane.SystemMicroservices.Nats.X86
			} else if controlPlane.SystemMicroservices.Nats.ARM != "" {
				images.Nats = controlPlane.SystemMicroservices.Nats.ARM
			}
			if images.Nats == "" {
				images.Nats = util.GetNatsImage()
			}
		} else {
			images.Nats = util.GetNatsImage()
		}
	}
	if images.Debugger == "" {
		// RemoteSystemMicroservices has no Debugger field; use util as fallback
		images.Debugger = util.GetDebuggerImage()
	}
}

// CollectControllerImages collects required images for controller deployment.
// When a controller already exists (!isInitialDeployment): catalog first, then YAML, then util.
// When no controller yet (isInitialDeployment): YAML then util (no catalog).
func CollectControllerImages(namespace string, controlPlane *rsc.RemoteControlPlane, isInitialDeployment bool) (*RequiredImages, error) {
	images := &RequiredImages{}

	// Controller image
	if controlPlane.Package.Container.Image != "" {
		images.Controller = controlPlane.Package.Container.Image
	} else {
		images.Controller = util.GetControllerImage()
	}

	if isInitialDeployment {
		// No controller to query; use YAML then util
		applyYAMLAndUtilFallbackForController(images, controlPlane)
		return images, nil
	}

	// Controller exists: try catalog first, then YAML, then util
	clt, err := clientutil.NewControllerClient(namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to controller: %w", err)
	}

	routerItem, err := getCatalogItemByName(clt, "router", "Router")
	if err == nil {
		applyRouterImagesFromCatalog(images, routerItem)
	}

	debuggerItem, err := getCatalogItemByName(clt, "debugger", "Debug")
	if err == nil {
		applyDebuggerImageFromCatalog(images, debuggerItem)
	} else {
		util.PrintNotify("Warning: Could not fetch debugger catalog item from controller. Debugger image will not be transferred.")
	}

	natsItem, err := getCatalogItemByName(clt, "nats", "NATS")
	if err == nil {
		applyNatsImageFromCatalog(images, natsItem)
	}

	applyYAMLAndUtilFallbackForController(images, controlPlane)
	return images, nil
}

// CollectAgentImages collects required images for agent deployment.
// When deploying an agent, a controller already exists; so we always try catalog first, then YAML, then util.
// controlPlane can be nil for Kubernetes or other non-remote control planes (catalog + util still apply).
func CollectAgentImages(namespace string, agent *rsc.RemoteAgent, controlPlane *rsc.RemoteControlPlane, _ bool) (*RequiredImages, error) {
	images := &RequiredImages{}

	// Agent image
	if agent.Package.Container.Image != "" {
		images.Agent = agent.Package.Container.Image
	} else {
		images.Agent = util.GetAgentImage()
	}

	// Router, NATS, debugger: catalog first (controller must exist when deploying agent), then YAML, then util
	clt, err := clientutil.NewControllerClient(namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to controller: %w", err)
	}

	routerItem, err := getCatalogItemByName(clt, "router", "Router")
	if err == nil {
		applyRouterImagesFromCatalog(images, routerItem)
	}

	debuggerItem, err := getCatalogItemByName(clt, "debugger", "Debug")
	if err == nil {
		applyDebuggerImageFromCatalog(images, debuggerItem)
	} else {
		util.PrintNotify("Warning: Could not fetch debugger catalog item from controller. Debugger image will not be transferred.")
	}

	natsItem, err := getCatalogItemByName(clt, "nats")
	if err == nil {
		applyNatsImageFromCatalog(images, natsItem)
	}

	applyYAMLAndUtilFallbackForAgent(images, controlPlane)
	return images, nil
}

// GetImageForPlatform returns the appropriate image based on platform
func GetImageForPlatform(images *RequiredImages, platform string) (string, error) {
	switch platform {
	case PlatformAMD64:
		return images.RouterX86, nil
	case PlatformARM64:
		return images.RouterARM, nil
	default:
		return "", util.NewInputError(fmt.Sprintf("unsupported platform %s", platform))
	}
}

// IsInitialDeployment checks if this is an initial control plane deployment
func IsInitialDeployment(namespace string) (bool, error) {
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return false, err
	}
	// If control plane exists and has controllers, this is not initial deployment
	controlPlane, err := ns.GetControlPlane()
	if err != nil {
		// No control plane exists, this is initial deployment
		return true, nil
	}
	// If controllers exist, this is not initial deployment
	controllers := controlPlane.GetControllers()
	return len(controllers) == 0, nil
}
