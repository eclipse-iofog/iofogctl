package deploycatalogitem

import (
	"fmt"

	apps "github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/apps"
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"gopkg.in/yaml.v2"
)

type Options struct {
	Namespace string
	Yaml      []byte
	Name      string
}

type remoteExecutor struct {
	catalogItem apps.CatalogItem
	namespace   string
}

func (exe *remoteExecutor) GetName() string {
	return exe.catalogItem.Name
}

func buildCatalogImages(item apps.CatalogItem) []client.CatalogImage {
	images := []client.CatalogImage{}
	if item.AMD64 != "" {
		images = append(images, client.CatalogImage{
			ContainerImage: item.AMD64,
			ArchID:         client.ArchNameToID["amd64"],
		})
	}
	if item.ARM64 != "" {
		images = append(images, client.CatalogImage{
			ContainerImage: item.ARM64,
			ArchID:         client.ArchNameToID["arm64"],
		})
	}
	if item.RISCV64 != "" {
		images = append(images, client.CatalogImage{
			ContainerImage: item.RISCV64,
			ArchID:         client.ArchNameToID["riscv64"],
		})
	}
	if item.ARM != "" {
		images = append(images, client.CatalogImage{
			ContainerImage: item.ARM,
			ArchID:         client.ArchNameToID["arm"],
		})
	}
	return images
}

func (exe *remoteExecutor) updateCatalogItem(clt *client.Client, existing *client.CatalogItemInfo) (err error) {
	request := client.CatalogItemUpdateRequest{
		ID:          existing.ID,
		Name:        exe.catalogItem.Name,
		Images:      buildCatalogImages(exe.catalogItem),
		Description: exe.catalogItem.Description,
	}

	if exe.catalogItem.Registry != "" {
		request.RegistryID, err = clientutil.ResolveRegistryID(exe.catalogItem.Registry)
		if err != nil {
			return err
		}
	}

	if _, err = clt.UpdateCatalogItem(&request); err != nil {
		return err
	}

	return nil
}

func (exe *remoteExecutor) createCatalogItem(clt *client.Client) (err error) {
	registryID, err := clientutil.ResolveRegistryID(exe.catalogItem.Registry)
	if err != nil {
		return err
	}

	if _, err = clt.CreateCatalogItem(&client.CatalogItemCreateRequest{
		Name:        exe.catalogItem.Name,
		Images:      buildCatalogImages(exe.catalogItem),
		RegistryID:  registryID,
		Description: exe.catalogItem.Description,
	}); err != nil {
		return err
	}
	return nil
}

func (exe *remoteExecutor) Execute() error {
	util.SpinStart(fmt.Sprintf("Deploying catalog item %s", exe.GetName()))
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	existing, err := clt.GetCatalogItemByName(exe.catalogItem.Name)
	if err != nil {
		if !clientutil.IsClientNotFoundError(err) {
			return err
		}
		return exe.createCatalogItem(clt)
	}
	return exe.updateCatalogItem(clt, existing)
}

func NewExecutor(opt Options) (exe execute.Executor, err error) {
	ns, err := config.GetNamespace(opt.Namespace)
	if err != nil {
		return exe, err
	}

	if len(ns.GetControllers()) == 0 {
		return exe, util.NewInputError("This namespace does not have a Controller. You must first deploy a Controller before deploying Applications")
	}

	var catalogItem apps.CatalogItem
	if err = yaml.UnmarshalStrict(opt.Yaml, &catalogItem); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}

	if len(opt.Name) > 0 {
		catalogItem.Name = opt.Name
	}

	if err := validate(&catalogItem); err != nil {
		return nil, err
	}

	return &remoteExecutor{
		namespace:   opt.Namespace,
		catalogItem: catalogItem,
	}, nil
}

func validate(opt *apps.CatalogItem) error {
	if opt.Name == "" {
		return util.NewInputError("Name must be specified")
	}
	if err := util.IsLowerAlphanumeric("CatalogItem", opt.Name); err != nil {
		return err
	}

	if opt.AMD64 == "" && opt.ARM64 == "" && opt.RISCV64 == "" && opt.ARM == "" {
		return util.NewInputError("At least one image must be specified")
	}

	if _, err := clientutil.ResolveRegistryID(opt.Registry); err != nil {
		return err
	}

	return nil
}
