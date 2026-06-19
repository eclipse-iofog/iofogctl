package install

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

// Remote agent uses SSH
type RemoteAgent struct {
	defaultAgent
	ssh     *util.SecureShellClient
	version string
	// repo          string
	// token         string
	dir           string
	procs         AgentProcedures
	customInstall bool // Flag set when custom install scripts are provided
	airgap        bool // Flag set when airgap deployment is enabled
}

type AgentProcedures struct {
	check          Entrypoint `yaml:"-"` // Check prereqs script (runs for default and custom procedures)
	Deps           Entrypoint `yaml:"deps,omitempty"`
	Install        Entrypoint `yaml:"install,omitempty"`
	Uninstall      Entrypoint `yaml:"uninstall,omitempty"`
	scriptNames    []string   `yaml:"-"` // List of all script names to be pushed to Agent
	scriptContents []string   `yaml:"-"` // List of contents of scripts to be pushed to Agent
}

type Entrypoint struct {
	Name     string   `yaml:"entrypoint"`
	Args     []string `yaml:"args"`
	destPath string   `yaml:"-"` // Dir + filename on Agent
}

func (script *Entrypoint) getCommand() string {
	args := strings.Join(script.Args, " ")
	return fmt.Sprintf("%s %s", script.destPath, args)
}

func NewRemoteAgent(user, host string, port int, privKeyFilename, agentName, agentUUID string) (*RemoteAgent, error) {
	ssh, err := util.NewSecureShellClient(user, host, privKeyFilename)
	if err != nil {
		return nil, err
	}
	ssh.SetPort(port)
	agent := &RemoteAgent{
		defaultAgent: defaultAgent{name: agentName, uuid: agentUUID},
		ssh:          ssh,
		version:      util.GetAgentVersion(),
		dir:          pkg.agentDir,
		procs: AgentProcedures{
			check: Entrypoint{
				Name:     pkg.scriptPrereq,
				destPath: util.JoinAgentPath(pkg.agentDir, pkg.scriptPrereq),
			},
			Deps: Entrypoint{
				Name:     pkg.scriptInstallDeps,
				destPath: util.JoinAgentPath(pkg.agentDir, pkg.scriptInstallDeps),
			},
			Install: Entrypoint{
				Name:     pkg.scriptInstallIofog,
				destPath: util.JoinAgentPath(pkg.agentDir, pkg.scriptInstallIofog),
				Args: []string{
					util.GetAgentVersion(),
					"",
					"",
				},
			},
			Uninstall: Entrypoint{
				Name:     pkg.scriptUninstallIofog,
				destPath: util.JoinAgentPath(pkg.agentDir, pkg.scriptUninstallIofog),
			},
			scriptNames: []string{
				pkg.scriptPrereq,
				pkg.scriptInit,
				pkg.scriptInstallDeps,
				pkg.scriptInstallJava,
				pkg.scriptInstallContainerEngine,
				pkg.scriptInstallIofog,
				pkg.scriptUninstallIofog,
			},
		},
	}
	// Get script contents from embedded files
	for _, scriptName := range agent.procs.scriptNames {
		scriptContent, err := util.GetStaticFile(addAgentAssetPrefix(scriptName))
		if err != nil {
			return nil, err
		}
		agent.procs.scriptContents = append(agent.procs.scriptContents, scriptContent)
	}
	return agent, nil
}

func NewRemoteContainerAgent(user, host string, port int, privKeyFilename, agentName, agentUUID, agentTZ string) (*RemoteAgent, error) {
	ssh, err := util.NewSecureShellClient(user, host, privKeyFilename)
	if err != nil {
		return nil, err
	}
	ssh.SetPort(port)
	if agentTZ == "" {
		agentTZ = "Europe/Istanbul"
	}
	agent := &RemoteAgent{
		defaultAgent: defaultAgent{name: agentName, uuid: agentUUID},
		ssh:          ssh,
		version:      util.GetAgentVersion(),
		dir:          pkg.agentDir,
		procs: AgentProcedures{
			check: Entrypoint{
				Name:     pkg.scriptPrereq,
				destPath: util.JoinAgentPath(pkg.agentDir, pkg.scriptPrereq),
			},
			Deps: Entrypoint{
				Name:     pkg.scriptInstallDeps,
				destPath: util.JoinAgentPath(pkg.agentDir, pkg.scriptInstallDeps),
			},
			Install: Entrypoint{
				Name:     pkg.scriptInstallIofog,
				destPath: util.JoinAgentPath(pkg.agentDir, pkg.scriptInstallIofog),
				Args: []string{
					util.GetAgentImage(),
					agentTZ,
					"",
				},
			},
			Uninstall: Entrypoint{
				Name:     pkg.scriptUninstallIofog,
				destPath: util.JoinAgentPath(pkg.agentDir, pkg.scriptUninstallIofog),
			},
			scriptNames: []string{
				pkg.scriptPrereq,
				pkg.scriptInit,
				pkg.scriptInstallDeps,
				pkg.scriptInstallContainerEngine,
				pkg.scriptInstallIofog,
				pkg.scriptUninstallIofog,
			},
		},
	}
	// Get script contents from embedded files
	for _, scriptName := range agent.procs.scriptNames {
		scriptContent, err := util.GetStaticFile(agent.addContainerAgentAssetPrefix(scriptName))
		if err != nil {
			return nil, err
		}
		agent.procs.scriptContents = append(agent.procs.scriptContents, scriptContent)
	}
	return agent, nil
}

func (agent *RemoteAgent) CustomizeProcedures(dir string, procs *AgentProcedures) error {
	// Format source directory of script files
	dir, err := util.FormatPath(dir)
	if err != nil {
		return err
	}

	// Load script files into memory
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, file := range files {
		if !file.IsDir() {
			procs.scriptNames = append(procs.scriptNames, file.Name())
			content, err := os.ReadFile(filepath.Join(dir, file.Name()))
			if err != nil {
				return err
			}
			procs.scriptContents = append(procs.scriptContents, string(content))
		}
	}

	// Add prereq script and entrypoint
	procs.scriptNames = append(procs.scriptNames, pkg.scriptPrereq)
	prereqContent, err := util.GetStaticFile(addAgentAssetPrefix(pkg.scriptPrereq))
	if err != nil {
		return err
	}
	procs.scriptContents = append(procs.scriptContents, prereqContent)
	procs.check.destPath = util.JoinAgentPath(agent.dir, pkg.scriptPrereq)

	// Add default entrypoints and scripts if necessary (user not provided)
	if procs.Deps.Name == "" {
		procs.Deps = agent.procs.Deps
		for _, script := range []string{pkg.scriptInstallDeps, pkg.scriptInstallContainerEngine, pkg.scriptInstallJava} {
			procs.scriptNames = append(procs.scriptNames, script)
			scriptContent, err := util.GetStaticFile(addAgentAssetPrefix(script))
			if err != nil {
				return err
			}
			procs.scriptContents = append(procs.scriptContents, scriptContent)
		}
	}
	if procs.Install.Name == "" {
		procs.Install = agent.procs.Install
		procs.scriptNames = append(procs.scriptNames, pkg.scriptInstallIofog)
		scriptContent, err := util.GetStaticFile(addAgentAssetPrefix(pkg.scriptInstallIofog))
		if err != nil {
			return err
		}
		procs.scriptContents = append(procs.scriptContents, scriptContent)
	} else {
		agent.customInstall = true
	}
	if procs.Uninstall.Name == "" {
		procs.Uninstall = agent.procs.Uninstall
		procs.scriptNames = append(procs.scriptNames, pkg.scriptUninstallIofog)
		scriptContent, err := util.GetStaticFile(addAgentAssetPrefix(pkg.scriptUninstallIofog))
		if err != nil {
			return err
		}
		procs.scriptContents = append(procs.scriptContents, scriptContent)
	}

	// Set destination paths where scripts appear on Agent
	procs.Deps.destPath = util.JoinAgentPath(agent.dir, procs.Deps.Name)
	procs.Install.destPath = util.JoinAgentPath(agent.dir, procs.Install.Name)
	procs.Uninstall.destPath = util.JoinAgentPath(agent.dir, procs.Uninstall.Name)

	agent.procs = *procs
	return nil
}

func (agent *RemoteAgent) SetVersion(version string) {
	if version == "" || agent.customInstall {
		return
	}
	agent.version = version
	agent.procs.Install.Args[0] = version
}

func (agent *RemoteAgent) SetContainerImage(image string) {
	if image == "" || agent.customInstall {
		return
	}
	agent.procs.Install.Args[0] = image
}

func (agent *RemoteAgent) SetAirgap(airgap bool) {
	agent.airgap = airgap
}

// func (agent *RemoteAgent) SetRepository(repo, token string) {
// 	if repo == "" || agent.customInstall {
// 		return
// 	}
// 	agent.repo = repo
// 	agent.procs.Install.Args[1] = repo
// 	agent.token = token
// 	agent.procs.Install.Args[2] = token
// }

func (agent *RemoteAgent) Bootstrap() error {
	// Prepare Agent for bootstrap
	if err := agent.copyInstallScriptsToAgent(); err != nil {
		return err
	}

	// Define bootstrap commands
	cmds := []command{
		{
			cmd: agent.procs.check.getCommand(),
			msg: "Checking prerequisites on Agent " + agent.name,
		},
		{
			cmd: agent.procs.Deps.getCommand(),
			msg: "Installing dependancies on Agent " + agent.name,
		},
		{
			cmd: fmt.Sprintf("sudo %s", agent.procs.Install.getCommand()),
			msg: "Installing ioFog daemon on Agent " + agent.name,
		},
	}

	// Execute commands on remote server
	if err := agent.run(cmds); err != nil {
		return err
	}

	return nil
}

func (agent *RemoteAgent) Configure(controllerEndpoint string, user IofogUser) (string, error) {
	key, caCert, err := agent.getProvisionKey(controllerEndpoint, user)
	if err != nil {
		return "", err
	}

	controllerBaseURL, err := util.GetBaseURL(controllerEndpoint)
	if err != nil {
		return "", err
	}
	// Instantiate commands
	cmds := []command{
		{
			cmd: "sudo iofog-agent config -a " + controllerBaseURL.String(),
			msg: "Configuring Agent " + agent.name + " with Controller URL " + controllerBaseURL.String(),
		},
	}

	// Only add cert command if caCert is not empty
	if caCert != "" {
		cmds = append(cmds, command{
			cmd: "sudo iofog-agent cert " + caCert,
			msg: "Configuring Agent " + agent.name + " with CA Certificate",
		})
	}

	cmds = append(cmds, command{
		cmd: "sudo iofog-agent provision " + key,
		msg: "Provisioning Agent " + agent.name + " with Controller",
	})

	// Execute commands on remote server
	if err := agent.run(cmds); err != nil {
		return "", err
	}

	return agent.uuid, nil
}

func (agent *RemoteAgent) SetInitialConfig(
	name, fogType string,
	// latitude, longitude float64,
	// description, fogType string,
	agentConfig client.AgentConfiguration,
) error {
	// Prepare the base commands for agent configuration
	cmds := []command{}

	// Convert FogType (string) to required format if necessary
	fogTypeValue := fogType
	if fogType == "" {
		fogTypeValue = "auto" // Default value if fogType is empty
	}

	// Convert WatchdogEnabled (*bool) to "on"/"off"
	watchdogEnabled := "off"
	if agentConfig.WatchdogEnabled != nil && *agentConfig.WatchdogEnabled {
		watchdogEnabled = "on"
	}

	// // Format GPS coordinates (Latitude and Longitude)
	// gpsCoordinates := ""
	// if latitude != 0 || longitude != 0 {
	// 	gpsCoordinates = fmt.Sprintf("%f,%f", latitude, longitude)
	// }

	// Extract values from agentConfig and construct options
	configOptions := map[string]string{
		"-ft": fogTypeValue,
		// "-gps": gpsCoordinates,
	}

	// Add values from agentConfig to configOptions, properly handling pointers
	if agentConfig.NetworkInterface != nil && *agentConfig.NetworkInterface != "" {
		configOptions["-n"] = *agentConfig.NetworkInterface
	}
	if agentConfig.DockerURL != nil && *agentConfig.DockerURL != "" {
		configOptions["-c"] = *agentConfig.DockerURL
	}
	if agentConfig.DiskLimit != nil {
		configOptions["-d"] = strconv.FormatInt(*agentConfig.DiskLimit, 10)
	}
	if agentConfig.DiskDirectory != nil && *agentConfig.DiskDirectory != "" {
		configOptions["-dl"] = *agentConfig.DiskDirectory
	}
	if agentConfig.MemoryLimit != nil {
		configOptions["-m"] = strconv.FormatInt(*agentConfig.MemoryLimit, 10)
	}
	if agentConfig.CPULimit != nil {
		configOptions["-p"] = strconv.FormatInt(*agentConfig.CPULimit, 10)
	}
	if agentConfig.LogLimit != nil {
		configOptions["-l"] = strconv.FormatInt(*agentConfig.LogLimit, 10)
	}
	if agentConfig.LogDirectory != nil && *agentConfig.LogDirectory != "" {
		configOptions["-ld"] = *agentConfig.LogDirectory
	}
	if agentConfig.LogFileCount != nil {
		configOptions["-lc"] = strconv.FormatInt(*agentConfig.LogFileCount, 10)
	}
	if agentConfig.StatusFrequency != nil {
		configOptions["-sf"] = strconv.FormatFloat(*agentConfig.StatusFrequency, 'f', -1, 64)
	}
	if agentConfig.ChangeFrequency != nil {
		configOptions["-cf"] = strconv.FormatFloat(*agentConfig.ChangeFrequency, 'f', -1, 64)
	}
	if agentConfig.DeviceScanFrequency != nil {
		configOptions["-sd"] = strconv.FormatFloat(*agentConfig.DeviceScanFrequency, 'f', -1, 64)
	}
	if agentConfig.LogLevel != nil && *agentConfig.LogLevel != "" {
		configOptions["-ll"] = *agentConfig.LogLevel
	}
	if agentConfig.AvailableDiskThreshold != nil {
		configOptions["-dt"] = strconv.FormatFloat(*agentConfig.AvailableDiskThreshold, 'f', -1, 64)
	}
	if agentConfig.DockerPruningFrequency != nil {
		configOptions["-pf"] = strconv.FormatFloat(*agentConfig.DockerPruningFrequency, 'f', -1, 64)
	}
	// if agentConfig.GpsDevice != nil && *agentConfig.GpsDevice != "" {
	// 	configOptions["-gpsd"] = *agentConfig.GpsDevice
	// }
	// if agentConfig.GpsMode != nil && *agentConfig.GpsMode != "" {
	// 	configOptions["-gps"] = *agentConfig.GpsMode
	// }
	// if agentConfig.GpsScanFrequency != nil {
	// 	configOptions["-gpsf"] = strconv.FormatFloat(*agentConfig.GpsScanFrequency, 'f', -1, 64)
	// }
	// if agentConfig.EdgeGuardFrequency != nil {
	// 	configOptions["-egf"] = strconv.FormatFloat(*agentConfig.EdgeGuardFrequency, 'f', -1, 64)
	// }
	if agentConfig.TimeZone != "" {
		configOptions["-tz"] = agentConfig.TimeZone
	}

	// Add watchdogEnabled to config options
	configOptions["-idc"] = watchdogEnabled

	// Iterate through the configOptions and add commands for non-empty values
	for option, value := range configOptions {
		if value != "" {
			cmds = append(cmds, command{
				cmd: fmt.Sprintf("sudo iofog-agent config %s %s", option, value),
				msg: fmt.Sprintf("Configuring Agent %s with option %s and value %s", name, option, value),
			})
		}
	}

	// If no commands were generated, return an error
	if len(cmds) == 0 {
		return fmt.Errorf("no valid configuration options provided for the agent")
	}

	// Execute commands on the remote server
	if err := agent.run(cmds); err != nil {
		return err
	}

	return nil
}

func (agent *RemoteAgent) Deprovision() (err error) {
	// Prepare commands
	cmds := []command{
		{
			cmd: "sudo iofog-agent deprovision",
			msg: "Deprovisioning Agent " + agent.name,
		},
	}

	// Execute commands on remote server
	if err = agent.run(cmds); err != nil && !isNotProvisionedError(err) {
		return
	}

	return
}

func (agent *RemoteAgent) Stop() (err error) {
	// Prepare commands
	cmds := []command{
		{
			cmd: "sudo iofog-agent deprovision",
			msg: "Deprovisioning Agent " + agent.name,
		},
	}
	if err = agent.run(cmds); err != nil && !isNotProvisionedError(err) {
		return err
	}

	cmds = []command{
		{
			cmd: "sudo -S service iofog-agent stop",
			msg: "Stopping Agent " + agent.name,
		},
	}
	if err := agent.run(cmds); err != nil {
		return err
	}

	return
}

func isNotProvisionedError(err error) bool {
	return strings.Contains(err.Error(), "not provisioned")
}

func (agent *RemoteAgent) Prune() (err error) {
	// Prepare commands
	cmds := []command{
		{
			// cmd: "sudo -S service iofog-agent prune",
			cmd: "sudo -S iofog-agent prune",
			msg: "Pruning Agent " + agent.name,
		},
	}

	// Execute commands on remote server
	if err := agent.run(cmds); err != nil {
		return err
	}

	return
}

func (agent *RemoteAgent) Uninstall() (err error) {
	// Stop iofog-agent properly
	if err = agent.Stop(); err != nil {
		return
	}

	// Prepare commands
	cmds := []command{
		// TODO: Implement purge on agent
		// {
		// 	cmd: "sudo iofog-agent purge",
		// 	msg: "Deprovisioning Agent " + agent.name,
		// },
		{
			cmd: agent.procs.Uninstall.getCommand(),
			msg: "Removing iofog-agent software " + agent.name,
		},
	}

	// Execute commands on remote server
	if err = agent.run(cmds); err != nil {
		return
	}

	return
}

func (agent *RemoteAgent) run(cmds []command) (err error) {
	// Establish SSH to agent
	if err = agent.ssh.Connect(); err != nil {
		return
	}
	defer util.Log(agent.ssh.Disconnect)

	// Execute commands
	for _, cmd := range cmds {
		Verbose(cmd.msg)
		if _, err = agent.ssh.Run(cmd.cmd); err != nil {
			return err
		}
	}

	return
}

func (agent *RemoteAgent) copyInstallScriptsToAgent() error {
	Verbose("Copying install scripts to Agent " + agent.name)
	cmds := []command{
		{
			cmd: fmt.Sprintf("sudo mkdir -p %s && sudo chmod -R 0777 %s", agent.dir, agent.dir),
			msg: "Creating Agent etc directory",
		},
	}
	if err := agent.run(cmds); err != nil {
		return err
	}
	return agent.copyScriptsToAgent()
}

func (agent *RemoteAgent) copyScriptsToAgent() error {
	// Establish SSH to agent
	if err := agent.ssh.Connect(); err != nil {
		return err
	}
	defer util.Log(agent.ssh.Disconnect)

	// Copy scripts to remote host
	for idx, script := range agent.procs.scriptNames {
		content := agent.procs.scriptContents[idx]
		reader := strings.NewReader(content)
		if err := agent.ssh.CopyTo(reader, agent.dir, script, "0775", int64(len(content))); err != nil {
			return err
		}
	}
	return nil
}

func addAgentAssetPrefix(file string) string {
	return fmt.Sprintf("agent/%s", file)
}

func (agent *RemoteAgent) addContainerAgentAssetPrefix(file string) string {
	if agent.airgap {
		return fmt.Sprintf("airgap-agent/%s", file)
	}
	return fmt.Sprintf("container-agent/%s", file)
}

type command struct {
	cmd string
	msg string
}
