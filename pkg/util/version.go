package util

import "fmt"

// Set by linker
var (
	versionNumber = "undefined"
	platform      = "undefined"
	commit        = "undefined"
	date          = "undefined"

	cliBinaryName   = "iofogctl"
	cliCrdGroup     = "iofog.org"
	cliApiVersion   = "iofog.org/v3"
	cliCpCrName     = "iofog"
	imageRegistry   = "ghcr.io/eclipse-iofog"
	cliDocsUrl      = "https://iofog.org"
	packageRepoBase = "https://iofog.datasance.com"
	ociSourceRepo   = "https://github.com/eclipse-iofog/iofogctl"

	controllerTag        = "undefined"
	operatorTag          = "undefined"
	routerTag            = "undefined"
	natsTag              = "undefined"
	edgeletTag           = "undefined"
	controllerVersion    = "undefined"
	edgeletVersion       = "undefined"
	edgeletReleaseBase   = "undefined"
	edgeletBinaryVersion = "undefined"
	edgeletGitHubRepo    = "eclipse-iofog/edgelet"
	debuggerTag          = "undefined"
)

const (
	controllerImage = "controller"
	edgeletImage    = "edgelet"
	operatorImage   = "operator"
	routerImage     = "router"
	routerARMImage  = "router"
	natsImage       = "nats"
	debuggerImage   = "node-debugger"
)

type Version struct {
	VersionNumber string `yaml:"version"`
	Platform      string
	Commit        string
	Date          string
}

func GetVersion() Version {
	return Version{
		VersionNumber: versionNumber,
		Platform:      platform,
		Commit:        commit,
		Date:          date,
	}
}

func GetCliBinaryName() string   { return cliBinaryName }
func GetCliCrdGroup() string     { return cliCrdGroup }
func GetCliApiVersion() string   { return cliApiVersion }
func GetCliCpCrName() string     { return cliCpCrName }
func GetImageRegistry() string   { return imageRegistry }
func GetCliDocsUrl() string      { return cliDocsUrl }
func GetPackageRepoBase() string { return packageRepoBase }
func GetOciSourceRepo() string   { return ociSourceRepo }

func GetControllerVersion() string    { return controllerVersion }
func GetEdgeletVersion() string       { return edgeletVersion }
func GetEdgeletReleaseBase() string   { return edgeletReleaseBase }
func GetEdgeletBinaryVersion() string { return edgeletBinaryVersion }
func GetEdgeletGitHubRepo() string    { return edgeletGitHubRepo }

func GetControllerImage() string {
	return fmt.Sprintf("%s/%s:%s", imageRegistry, controllerImage, controllerTag)
}
func GetEdgeletImage() string {
	return fmt.Sprintf("%s/%s:%s", imageRegistry, edgeletImage, edgeletTag)
}
func GetOperatorImage() string {
	return fmt.Sprintf("%s/%s:%s", imageRegistry, operatorImage, operatorTag)
}
func GetRouterImage() string {
	return fmt.Sprintf("%s/%s:%s", imageRegistry, routerImage, routerTag)
}
func GetNatsImage() string {
	return fmt.Sprintf("%s/%s:%s", imageRegistry, natsImage, natsTag)
}
func GetDebuggerImage() string {
	return fmt.Sprintf("%s/%s:%s", imageRegistry, debuggerImage, debuggerTag)
}

// Deprecated: Compatibility wrappers for Phase 1 compilation. Will be removed in Phase 4/5.
func GetAgentImage() string   { return GetEdgeletImage() }
func GetAgentVersion() string { return GetEdgeletVersion() }
