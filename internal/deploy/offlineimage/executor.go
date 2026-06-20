package deployofflineimage

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"gopkg.in/yaml.v2"
)

// Options represents the executor input.
type Options struct {
	Namespace string
	Name      string
	Yaml      []byte
	NoCache   bool
	PoolSize  int
}

type executor struct {
	namespace string
	spec      rsc.OfflineImage
	noCache   bool
	poolSize  int
}

const offlineRegistryID = 2

func (exe *executor) GetName() string {
	return exe.spec.Name
}

// NewExecutor builds an OfflineImage executor from options.
func NewExecutor(opt Options) (execute.Executor, error) {
	var definition rsc.OfflineImage
	if err := yaml.UnmarshalStrict(opt.Yaml, &definition); err != nil {
		return nil, util.NewUnmarshalError(err.Error())
	}
	if opt.Name != "" {
		definition.Name = opt.Name
	}
	if err := validateDefinition(&definition); err != nil {
		return nil, err
	}

	return &executor{
		namespace: opt.Namespace,
		spec:      definition,
		noCache:   opt.NoCache,
		poolSize:  sanitizePoolSize(opt.PoolSize),
	}, nil
}

func sanitizePoolSize(size int) int {
	if size <= 0 {
		return 2
	}
	return size
}

func (exe *executor) Execute() error {
	defer util.SpinStop()

	util.SpinStart(fmt.Sprintf("Validating OfflineImage %s", exe.spec.Name))
	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	if len(ns.GetControllers()) == 0 {
		return util.NewInputError("This namespace does not have a Controller. Deploy a Controller before deploying OfflineImages")
	}

	plans, err := exe.buildAgentPlans(ns)
	if err != nil {
		return err
	}
	if len(plans) == 0 {
		return util.NewInputError(fmt.Sprintf("OfflineImage %s does not target any agents", exe.spec.Name))
	}

	util.SpinStart("Pulling offline images for target platforms")
	artifacts, err := exe.prepareArtifacts(plans)
	if err != nil {
		return err
	}
	defer cleanupArtifacts(artifacts)

	util.SpinStart("Transferring offline images to agents")
	if err := exe.distribute(plans, artifacts); err != nil {
		return err
	}

	util.SpinStart("Registering offline catalog item")
	if err := exe.registerCatalogItem(); err != nil {
		return err
	}

	util.SpinStart(fmt.Sprintf("OfflineImage %s deployment completed", exe.spec.Name))
	return nil
}

func validateDefinition(def *rsc.OfflineImage) error {
	if def.Name == "" {
		return util.NewInputError("OfflineImage spec must include a name")
	}
	if err := util.IsLowerAlphanumeric("OfflineImage", def.Name); err != nil {
		return err
	}
	if len(def.Agents) == 0 {
		return util.NewInputError("OfflineImage spec must include at least one agent entry")
	}
	if def.X86Image == "" && def.ArmImage == "" {
		return util.NewInputError("OfflineImage spec must include at least one architecture image (x86 or arm)")
	}
	if def.Auth != nil {
		if def.Auth.Username == "" || def.Auth.Password == "" {
			return util.NewInputError("OfflineImage auth requires both username and password when provided")
		}
	}
	return nil
}

func (exe *executor) buildAgentPlans(ns *rsc.Namespace) ([]agentPlan, error) {
	plans := make([]agentPlan, 0, len(exe.spec.Agents))
	for _, agentName := range exe.spec.Agents {
		baseAgent, err := ns.GetAgent(agentName)
		if err != nil {
			return nil, err
		}
		remoteAgent, ok := baseAgent.(*rsc.RemoteAgent)
		if !ok {
			return nil, util.NewInputError(fmt.Sprintf("OfflineImage only supports remote Agents. %s is not a remote Agent", agentName))
		}
		if err := remoteAgent.ValidateSSH(); err != nil {
			return nil, err
		}
		cfg, _, _, err := clientutil.GetAgentConfig(agentName, exe.namespace)
		if err != nil {
			return nil, err
		}
		platform, err := resolvePlatform(cfg.FogType)
		if err != nil {
			return nil, fmt.Errorf("agent %s: %w", agentName, err)
		}
		imageRef, err := exe.imageForPlatform(platform)
		if err != nil {
			return nil, fmt.Errorf("agent %s: %w", agentName, err)
		}
		engine, err := resolveContainerEngine(cfg.ContainerEngine)
		if err != nil {
			return nil, fmt.Errorf("agent %s: %w", agentName, err)
		}
		plans = append(plans, agentPlan{
			agent:    remoteAgent,
			platform: platform,
			engine:   engine,
			imageRef: imageRef,
		})
	}
	return plans, nil
}

func (exe *executor) imageForPlatform(platform string) (string, error) {
	switch platform {
	case platformAMD64:
		if exe.spec.X86Image == "" {
			return "", util.NewInputError("x86 image is required for agents with linux/amd64 fog type")
		}
		return exe.spec.X86Image, nil
	case platformARM64:
		if exe.spec.ArmImage == "" {
			return "", util.NewInputError("arm image is required for agents with linux/arm64 fog type")
		}
		return exe.spec.ArmImage, nil
	default:
		return "", util.NewInputError(fmt.Sprintf("unsupported platform %s", platform))
	}
}

func (exe *executor) prepareArtifacts(plans []agentPlan) (map[string]*imageArtifact, error) {
	ctx := context.Background()
	artifacts := make(map[string]*imageArtifact)
	for platform, imageRef := range collectPlatformImages(plans) {
		artifact, err := exe.ensureArtifact(ctx, platform, imageRef)
		if err != nil {
			return nil, err
		}
		artifacts[platform] = artifact
	}
	return artifacts, nil
}

func collectPlatformImages(plans []agentPlan) map[string]string {
	unique := make(map[string]string)
	for _, plan := range plans {
		if _, exists := unique[plan.platform]; exists {
			continue
		}
		unique[plan.platform] = plan.imageRef
	}
	return unique
}

func (exe *executor) registerCatalogItem() error {
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	images := []client.CatalogImage{}
	if exe.spec.X86Image != "" {
		images = append(images, client.CatalogImage{
			ContainerImage: exe.spec.X86Image,
			AgentTypeID:    client.AgentTypeAgentTypeIDDict["x86"],
		})
	}
	if exe.spec.ArmImage != "" {
		images = append(images, client.CatalogImage{
			ContainerImage: exe.spec.ArmImage,
			AgentTypeID:    client.AgentTypeAgentTypeIDDict["arm"],
		})
	}
	item, err := clt.GetCatalogItemByName(exe.spec.Name)
	if err != nil {
		if !isNotFoundError(err) {
			return err
		}
		_, err = clt.CreateCatalogItem(&client.CatalogItemCreateRequest{
			Name:        exe.spec.Name,
			Images:      images,
			RegistryID:  offlineRegistryID,
			Description: fmt.Sprintf("Offline image bundle %s", exe.spec.Name),
		})
		return err
	}

	proceed, err := exe.confirmCatalogUpdate(item)
	if err != nil {
		return err
	}
	if !proceed {
		util.PrintInfo("Catalog item update skipped by user request")
		return nil
	}

	_, err = clt.UpdateCatalogItem(&client.CatalogItemUpdateRequest{
		ID:          item.ID,
		Name:        exe.spec.Name,
		Images:      images,
		RegistryID:  offlineRegistryID,
		Description: fmt.Sprintf("Offline image bundle %s", exe.spec.Name),
	})
	return err
}

func (exe *executor) confirmCatalogUpdate(item *client.CatalogItemInfo) (bool, error) {
	util.SpinHandlePrompt()
	defer util.SpinHandlePromptComplete()

	imageDetails := "Images: none"
	if len(item.Images) > 0 {
		segments := make([]string, 0, len(item.Images))
		for _, img := range item.Images {
			segments = append(segments, fmt.Sprintf("%s (AgentTypeID: %d)", img.ContainerImage, img.AgentTypeID))
		}
		imageDetails = strings.Join(segments, " | ")
	}
	fmt.Printf("\nCatalog item '%s' already exists (ID: %d).\nRegistry ID: %d | %s | Category: %s\n", item.Name, item.ID, item.RegistryID, imageDetails, item.Category)
	fmt.Println("Updating this item will immediately affect every microservice referencing it.")
	if strings.EqualFold(item.Category, "SYSTEM") {
		fmt.Println("WARNING: Category is SYSTEM. Changes may impact critical system microservices.")
	}
	if item.RegistryID != offlineRegistryID {
		fmt.Println("WARNING: Existing catalog item is bound to an online registry (RegistryID != 2). Ensure agents are prepared for this change.")
	}
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Proceed with updating the catalog item? (y/n): ")
		resp, err := reader.ReadString('\n')
		if err != nil {
			return false, fmt.Errorf("failed to read user response: %w", err)
		}
		resp = strings.TrimSpace(resp)
		switch strings.ToLower(resp) {
		case "y":
			return true, nil
		case "n":
			return false, nil
		default:
			fmt.Println("Please enter 'y' or 'n'.")
		}
	}
}

func (exe *executor) distribute(plans []agentPlan, artifacts map[string]*imageArtifact) error {
	var wg sync.WaitGroup
	type transferResult struct {
		agent string
		err   error
	}
	results := make(chan transferResult, len(plans))
	poolSize := exe.poolSize
	if poolSize <= 0 {
		poolSize = 1
	}
	sem := make(chan struct{}, poolSize)
	for _, plan := range plans {
		plan := plan
		artifact := artifacts[plan.platform]
		if artifact == nil {
			results <- transferResult{agent: plan.agent.Name, err: fmt.Errorf("missing artifact for platform %s", plan.platform)}
			continue
		}
		wg.Add(1)
		go func(plan agentPlan, artifact *imageArtifact) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			if err := transferArtifact(plan, artifact); err != nil {
				results <- transferResult{agent: plan.agent.Name, err: err}
				return
			}
			results <- transferResult{agent: plan.agent.Name, err: nil}
		}(plan, artifact)
	}
	wg.Wait()
	close(results)

	failures := []transferResult{}
	successes := []string{}
	for res := range results {
		if res.err != nil {
			failures = append(failures, res)
		} else {
			successes = append(successes, res.agent)
		}
	}

	util.PrintInfo(fmt.Sprintf("Offline image transfer summary: %d success, %d failed", len(successes), len(failures)))
	if len(successes) > 0 {
		util.PrintInfo("Successful agents:")
		for _, agent := range successes {
			fmt.Printf("  - %s\n", agent)
		}
		fmt.Println()
	}
	if len(failures) > 0 {
		util.PrintNotify("Failed agents:")
		for _, fail := range failures {
			fmt.Printf("  - %s: %s\n", fail.agent, fail.err)
		}
	}

	return nil
}
