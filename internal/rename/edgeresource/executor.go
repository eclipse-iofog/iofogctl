package edgeresource

import (
	"fmt"

	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func Execute(namespace, name, newName string) error {
	if err := util.IsLowerAlphanumeric("Edge Resource", newName); err != nil {
		return err
	}

	util.SpinStart(fmt.Sprintf("Renaming edgeResource %s", name))

	// Init remote resources
	clt, err := clientutil.NewControllerClient(namespace)
	if err != nil {
		return err
	}

	// List all edge resources
	listResponse, err := clt.ListEdgeResources()
	if err != nil {
		return err
	}
	// Validate exists
	if len(listResponse.EdgeResources) == 0 {
		return util.NewNotFoundError(fmt.Sprintf("%s does not exist", name))
	}

	// Get full resource contents and update
	for idx := range listResponse.EdgeResources {
		meta := &listResponse.EdgeResources[idx]
		if meta.Name != name {
			continue
		}
		// Get versioned resource
		oldEdge, err := clt.GetHTTPEdgeResourceByName(meta.Name, meta.Version)
		if err != nil {
			return err
		}
		// Update versioned resource
		oldEdge.Name = newName
		if err := clt.UpdateHTTPEdgeResource(name, &oldEdge); err != nil {
			return err
		}
	}

	return nil
}
