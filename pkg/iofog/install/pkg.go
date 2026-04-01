/*
 *  *******************************************************************************
 *  * Copyright (c) 2023 Contributors to the Eclipse ioFog Project
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package install

var pkg struct {
	scriptPrereq                           string
	scriptInit                             string
	scriptInstallDeps                      string
	scriptInstallJava                      string
	scriptInstallContainerEngine           string
	scriptInstallIofog                     string
	scriptUninstallIofog                   string
	controllerScriptPrereq                 string
	controllerScriptInit                   string
	controllerScriptInstallContainerEngine string
	controllerScriptSetEnv                 string
	controllerScriptInstall                string
	controllerScriptUninstall              string
	iofogDir                               string
	agentDir                               string
	controllerDir                          string
}

func init() {
	pkg.scriptPrereq = "check_prereqs.sh"
	pkg.scriptInit = "init.sh"
	pkg.scriptInstallDeps = "install_deps.sh"
	pkg.scriptInstallJava = "install_java.sh"
	pkg.scriptInstallContainerEngine = "install_container_engine.sh"
	pkg.scriptInstallIofog = "install_iofog.sh"
	pkg.scriptUninstallIofog = "uninstall_iofog.sh"
	pkg.controllerScriptPrereq = "check_prereqs.sh"
	pkg.controllerScriptInit = "init.sh"
	pkg.controllerScriptInstallContainerEngine = "install_container_engine.sh"
	pkg.controllerScriptSetEnv = "set_env.sh"
	pkg.controllerScriptInstall = "install_iofog.sh"
	pkg.controllerScriptUninstall = "uninstall_iofog.sh"
	pkg.iofogDir = "/etc/iofog"
	pkg.agentDir = "/etc/iofog/agent"
	pkg.controllerDir = "/etc/iofog/controller"
}
