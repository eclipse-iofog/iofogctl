package install

var pkg struct {
	edgeletScriptPrereq                    string
	edgeletScriptDetectInit                string
	edgeletScriptInstallDeps               string
	edgeletScriptConfigureContainerEngine  string
	edgeletScriptInstall                   string
	edgeletScriptInstallContainer          string
	edgeletScriptInstallInitUnits          string
	edgeletScriptStartEdgelet              string
	edgeletScriptConfigureContainerEdgelet string
	edgeletScriptWaitEdgeletReady          string
	edgeletScriptBundled                   string
	edgeletScriptUninstall                 string
	edgeletLibScripts                      []string
	controllerScriptPrereq                 string
	controllerScriptInit                   string
	controllerScriptInstallContainerEngine string
	controllerScriptSetEnv                 string
	controllerScriptInstall                string
	controllerScriptUninstall              string
	iofogDir                               string
	controllerDir                          string
}

func init() {
	pkg.edgeletScriptPrereq = "check_prereqs.sh"
	pkg.edgeletScriptDetectInit = "detect_init.sh"
	pkg.edgeletScriptInstallDeps = "install_deps.sh"
	pkg.edgeletScriptConfigureContainerEngine = "configure_container_engine.sh"
	pkg.edgeletScriptInstall = "install.sh"
	pkg.edgeletScriptInstallContainer = "install_container.sh"
	pkg.edgeletScriptInstallInitUnits = "install_init_units.sh"
	pkg.edgeletScriptStartEdgelet = "start_edgelet.sh"
	pkg.edgeletScriptConfigureContainerEdgelet = "configure_container_edgelet.sh"
	pkg.edgeletScriptWaitEdgeletReady = "wait_edgelet_ready.sh"
	pkg.edgeletScriptBundled = "bundled.sh"
	pkg.edgeletScriptUninstall = "uninstall.sh"
	pkg.edgeletLibScripts = []string{
		"lib/common.sh",
		"lib/paths.sh",
		"lib/receipt.sh",
		"lib/binary.sh",
		"lib/container_cli.sh",
		"lib/container_engine.sh",
		"lib/container_mounts.sh",
	}
	pkg.controllerScriptPrereq = "check_prereqs.sh"
	pkg.controllerScriptInit = "init.sh"
	pkg.controllerScriptInstallContainerEngine = "install_container_engine.sh"
	pkg.controllerScriptSetEnv = "set_env.sh"
	pkg.controllerScriptInstall = "install_iofog.sh"
	pkg.controllerScriptUninstall = "uninstall_iofog.sh"
	pkg.iofogDir = "/etc/iofog"
	pkg.controllerDir = "/etc/iofog/controller"
}
