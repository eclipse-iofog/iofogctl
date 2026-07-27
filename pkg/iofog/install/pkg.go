package install

var pkg struct {
	edgeletScriptPrereq                    string
	edgeletScriptDetectInit                string
	edgeletScriptInstallDeps               string
	edgeletScriptConfigureContainerEngine  string
	edgeletScriptInstall                   string
	edgeletScriptInstallWasmRuntimes       string
	edgeletScriptInstallContainer          string
	edgeletScriptInstallInitUnits          string
	edgeletScriptStartEdgelet              string
	edgeletScriptProbeContainerdReady      string
	edgeletScriptProbeEdgeletReady         string
	edgeletScriptRestartContainerd         string
	edgeletScriptConfigureContainerEdgelet string
	edgeletScriptWaitEdgeletReady          string
	edgeletScriptBundled                   string
	edgeletScriptUninstall                 string
	edgeletLibScripts                      []string
	iofogDir                               string
}

func init() {
	pkg.edgeletScriptPrereq = "check_prereqs.sh"
	pkg.edgeletScriptDetectInit = "detect_init.sh"
	pkg.edgeletScriptInstallDeps = "install_deps.sh"
	pkg.edgeletScriptConfigureContainerEngine = "configure_container_engine.sh"
	pkg.edgeletScriptInstall = "install.sh"
	pkg.edgeletScriptInstallWasmRuntimes = "install_wasm_runtimes.sh"
	pkg.edgeletScriptInstallContainer = "install_container.sh"
	pkg.edgeletScriptInstallInitUnits = "install_init_units.sh"
	pkg.edgeletScriptStartEdgelet = "start_edgelet.sh"
	pkg.edgeletScriptProbeContainerdReady = "probe_containerd_ready.sh"
	pkg.edgeletScriptProbeEdgeletReady = "probe_edgelet_ready.sh"
	pkg.edgeletScriptRestartContainerd = "restart_edgelet_containerd.sh"
	pkg.edgeletScriptConfigureContainerEdgelet = "configure_container_edgelet.sh"
	pkg.edgeletScriptWaitEdgeletReady = "wait_edgelet_ready.sh"
	pkg.edgeletScriptBundled = "bundled.sh"
	pkg.edgeletScriptUninstall = "uninstall.sh"
	pkg.edgeletLibScripts = []string{
		"lib/common.sh",
		"lib/paths.sh",
		"lib/receipt.sh",
		"lib/binary.sh",
		"lib/embed.sh",
		"lib/service.sh",
		"lib/container_cli.sh",
		"lib/container_engine.sh",
		"lib/container_mounts.sh",
	}
	pkg.iofogDir = "/etc/iofog"
}
