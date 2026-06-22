package application

import (
	"fmt"

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

	_, err = clt.GetApplicationByName(name)
	if err != nil {
		return err
	}

	// application.Name = newName
	// // _, err = clt.UpdateApplication(&client.ApplicationUpdateRequest{
	// // 	ID:   application.ID,
	// // 	Name: &newName,
	// // })
	// fmt.Println(application)
	return fmt.Errorf("Application renamed not allowed")
}
