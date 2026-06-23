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
	Controller      string
	Agent           string
	RouterAMD64     string
	RouterARM64     string
	RouterRISCV64   string
	RouterARM       string
	NatsAMD64       string
	NatsARM64       string
	NatsRISCV64     string
	NatsARM         string
	DebuggerAMD64   string
	DebuggerARM64   string
	DebuggerRISCV64 string
	DebuggerARM     string
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
		switch client.ArchIDToName[img.ArchID] {
		case "amd64":
			images.RouterAMD64 = img.ContainerImage
		case "arm64":
			images.RouterARM64 = img.ContainerImage
		case "riscv64":
			images.RouterRISCV64 = img.ContainerImage
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
		switch client.ArchIDToName[img.ArchID] {
		case "amd64":
			images.DebuggerAMD64 = img.ContainerImage
		case "arm64":
			images.DebuggerARM64 = img.ContainerImage
		case "riscv64":
			images.DebuggerRISCV64 = img.ContainerImage
		case "arm":
			images.DebuggerARM = img.ContainerImage
		}
	}
}

// applyNatsImageFromCatalog sets Nats from catalog item (first image).
func applyNatsImageFromCatalog(images *RequiredImages, item *client.CatalogItemInfo) {
	if item == nil || len(item.Images) == 0 {
		return
	}
	for _, img := range item.Images {
		switch client.ArchIDToName[img.ArchID] {
		case "amd64":
			images.NatsAMD64 = img.ContainerImage
		case "arm64":
			images.NatsARM64 = img.ContainerImage
		case "riscv64":
			images.NatsRISCV64 = img.ContainerImage
		case "arm":
			images.NatsARM = img.ContainerImage
		}
	}
}

// applyYAMLFallbackForController fills any empty router/nats/debugger from controlPlane.SystemMicroservices; util is last fallback.
func applyYAMLAndUtilFallbackForController(images *RequiredImages, controlPlane *rsc.RemoteControlPlane) {
	if controlPlane == nil {
		return
	}
	sm := &controlPlane.SystemMicroservices
	if images.RouterAMD64 == "" {
		if sm.Router.AMD64 != "" {
			images.RouterAMD64 = sm.Router.AMD64
		} else {
			images.RouterAMD64 = util.GetRouterImage()
		}
	}
	if images.RouterARM64 == "" {
		if sm.Router.ARM != "" {
			images.RouterARM64 = sm.Router.ARM64
		} else {
			images.RouterARM64 = util.GetRouterImage()
		}
	}
	if images.RouterRISCV64 == "" {
		if sm.Router.RISCV64 != "" {
			images.RouterRISCV64 = sm.Router.RISCV64
		} else {
			images.RouterRISCV64 = util.GetRouterImage()
		}
	}
	if images.RouterARM == "" {
		if sm.Router.ARM != "" {
			images.RouterARM = sm.Router.ARM
		} else {
			images.RouterARM = util.GetRouterImage()
		}
	}
	if images.NatsAMD64 == "" {
		if sm.Nats.AMD64 != "" {
			images.NatsAMD64 = sm.Nats.AMD64
		} else if sm.Nats.ARM64 != "" {
			images.NatsARM64 = sm.Nats.ARM64
		} else {
			images.NatsAMD64 = util.GetNatsImage()
		}
	}
	if images.NatsARM64 == "" {
		if sm.Nats.ARM64 != "" {
			images.NatsARM64 = sm.Nats.ARM64
		} else {
			images.NatsARM64 = util.GetNatsImage()
		}
	}
	if images.NatsRISCV64 == "" {
		if sm.Nats.RISCV64 != "" {
			images.NatsRISCV64 = sm.Nats.RISCV64
		} else {
			images.NatsRISCV64 = util.GetNatsImage()
		}
	}
	if images.NatsARM == "" {
		if sm.Nats.ARM != "" {
			images.NatsARM = sm.Nats.ARM
		} else {
			images.NatsARM = util.GetNatsImage()
		}
	}
	if images.DebuggerAMD64 == "" {
		if images.DebuggerAMD64 != "" {
			images.DebuggerAMD64 = images.DebuggerAMD64
		} else {
			images.DebuggerAMD64 = util.GetDebuggerImage()
		}
	}
	if images.DebuggerARM64 == "" {
		if images.DebuggerARM64 != "" {
			images.DebuggerARM64 = images.DebuggerARM64
		} else {
			images.DebuggerARM64 = util.GetDebuggerImage()
		}
	}
	if images.DebuggerRISCV64 == "" {
		if images.DebuggerRISCV64 != "" {
			images.DebuggerRISCV64 = images.DebuggerRISCV64
		} else {
			images.DebuggerRISCV64 = util.GetDebuggerImage()
		}
	}
	if images.DebuggerARM == "" {
		if images.DebuggerARM != "" {
			images.DebuggerARM = images.DebuggerARM
		} else {
			images.DebuggerARM = util.GetDebuggerImage()
		}
	}
}

// applyYAMLAndUtilFallbackForAgent fills any empty router/nats/debugger from controlPlane (if non-nil) then util.
func applyYAMLAndUtilFallbackForAgent(images *RequiredImages, controlPlane *rsc.RemoteControlPlane) {
	if images.RouterAMD64 == "" {
		if controlPlane != nil && controlPlane.SystemMicroservices.Router.AMD64 != "" {
			images.RouterAMD64 = controlPlane.SystemMicroservices.Router.AMD64
		} else {
			images.RouterAMD64 = util.GetRouterImage()
		}
	}
	if images.RouterARM64 == "" {
		if controlPlane != nil && controlPlane.SystemMicroservices.Router.ARM64 != "" {
			images.RouterARM64 = controlPlane.SystemMicroservices.Router.ARM64
		} else {
			images.RouterARM64 = util.GetRouterImage()
		}
	}
	if images.NatsAMD64 == "" {
		if controlPlane != nil {
			if controlPlane.SystemMicroservices.Nats.AMD64 != "" {
				images.NatsAMD64 = controlPlane.SystemMicroservices.Nats.AMD64
			} else if controlPlane.SystemMicroservices.Nats.ARM != "" {
				images.NatsARM64 = controlPlane.SystemMicroservices.Nats.ARM64
			}
			if images.NatsAMD64 == "" {
				images.NatsAMD64 = util.GetNatsImage()
			}
		} else {
			images.NatsAMD64 = util.GetNatsImage()
		}
	}
	if images.DebuggerAMD64 == "" {
		// RemoteSystemMicroservices has no Debugger field; use util as fallback
		images.DebuggerAMD64 = util.GetDebuggerImage()
	}
	if images.DebuggerARM64 == "" {
		if images.DebuggerARM64 != "" {
			images.DebuggerARM64 = images.DebuggerARM64
		} else {
			images.DebuggerARM64 = util.GetDebuggerImage()
		}
	}
	if images.DebuggerRISCV64 == "" {
		if images.DebuggerRISCV64 != "" {
			images.DebuggerRISCV64 = images.DebuggerRISCV64
		} else {
			images.DebuggerRISCV64 = util.GetDebuggerImage()
		}
	}
	if images.DebuggerARM == "" {
		if images.DebuggerARM != "" {
			images.DebuggerARM = images.DebuggerARM
		} else {
			images.DebuggerARM = util.GetDebuggerImage()
		}
	}
}

// CollectControllerImages collects required images for controller deployment.
// When a controller already exists (!isInitialDeployment): catalog first, then YAML, then util.
// When no controller yet (isInitialDeployment): YAML then util (no catalog).
func CollectControllerImages(namespace string, controlPlane *rsc.RemoteControlPlane, isInitialDeployment bool) (*RequiredImages, error) {
	images := &RequiredImages{}

	// Controller image
	if controlPlane.Controller.Package != nil && controlPlane.Controller.Package.Image != "" {
		images.Controller = controlPlane.Controller.Package.Image
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
		return images.RouterAMD64, nil
	case PlatformARM64:
		return images.RouterARM64, nil
	case PlatformRISCV64:
		return images.RouterRISCV64, nil
	case PlatformARM:
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
