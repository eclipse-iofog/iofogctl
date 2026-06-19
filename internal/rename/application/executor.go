package application

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func Execute(namespace, name, newName string) error {
	if err := util.IsLowerAlphanumeric("Application", newName); err != nil {
		return err
	}
	util.SpinStart(fmt.Sprintf("Renaming Application %s", name))

	// Init remote resources
	clt, err := clientutil.NewControllerClient(namespace)
	if err != nil {
		return err
	}

	flow, err := clt.GetFlowByName(name)
	if err != nil {
		return err
	}

	flow.Name = newName
	_, err = clt.UpdateFlow(&client.FlowUpdateRequest{
		ID:   flow.ID,
		Name: &newName,
	})
	if err != nil {
		return err
	}
	config.Flush()
	return nil
}
