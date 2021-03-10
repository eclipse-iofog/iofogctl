/*
 *  *******************************************************************************
 *  * Copyright (c) 2020 Edgeworx, Inc.
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package cmd

import (
	"context"
	"fmt"

	"github.com/eclipse-iofog/iofogctl/v3/pkg/iofog/install"

	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	clientutil "github.com/eclipse-iofog/iofogctl/v3/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // GCP Auth pkg required
	"k8s.io/client-go/tools/clientcmd"
)

const (
	sshErrMsg = "legacy commands requires SSH access.\n%s %s SSH details are not available.\nUse `iofogctl configure --help` to find out how to add SSH details"
)

func k8sExecute(kubeConfig, namespace, podSelector string, cliCmd, cmd []string) {
	kubeConfig, err := util.FormatPath(kubeConfig)
	util.Check(err)
	// Connect to cluster
	// Execute
	conf, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	util.Check(err)
	// Instantiate Kubernetes client
	clientset, err := kubernetes.NewForConfig(conf)
	util.Check(err)
	podList, err := clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{LabelSelector: podSelector})
	if err != nil {
		return
	}
	podName := podList.Items[0].Name
	kubeArgs := []string{"exec", podName, "-n", namespace, "--"}
	kubeArgs = append(kubeArgs, cliCmd...)
	kubeArgs = append(kubeArgs, cmd...)
	kubectlCmd := "kubectl"
	for _, kArg := range kubeArgs {
		kubectlCmd = kubectlCmd + " " + kArg
	}
	util.PrintNotify("Cannot use legacy command against a Kubernetes Controller. Use the following command instead: \n\n  " + kubectlCmd)
}

func localExecute(container string, localCLI, localCmd []string) {
	// Execute command
	localContainerClient, err := install.NewLocalContainerClient()
	util.Check(err)
	cmd := append(localCLI, localCmd...)
	result, err := localContainerClient.ExecuteCmd(container, cmd)
	util.Check(err)
	fmt.Print(result.StdOut)
	if len(result.StdErr) > 0 {
		util.PrintError(result.StdErr)
	}
}

func remoteExec(user, host, keyFile string, port int, cliCmd string, cmd []string) {
	ssh, err := util.NewSecureShellClient(user, host, keyFile)
	util.Check(err)
	ssh.SetPort(port)
	util.Check(ssh.Connect())
	defer util.Log(ssh.Disconnect)

	sshCmd := cliCmd
	for _, arg := range cmd {
		sshCmd = sshCmd + " " + arg
	}
	logs, err := ssh.Run(sshCmd)
	util.Check(err)
	fmt.Print(logs.String())
}

// NOTE: (Serge) This code will be discarded eventually. Keeping it one file.
func newLegacyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "legacy resource NAME COMMAND ARGS...",
		Short: "Execute commands using legacy CLI",
		Long: `Execute commands using legacy Controller and Agent CLI.

Legacy commands require SSH access to the corresponding Agent or Controller.

Use the configure command to add SSH details to Agents and Controllers if necessary.`,
		Example: `iofogctl legacy controller NAME COMMAND
iofogctl legacy agent      NAME COMMAND`,
		Args: cobra.MinimumNArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			// Get resource type arg
			resource := args[0]
			// Get resource name
			name := args[1]
			// Get namespace
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)
			useDetached, err := cmd.Flags().GetBool("detached")
			util.Check(err)

			ns, err := config.GetNamespace(namespace)
			util.Check(err)
			switch resource {
			case "controller":
				// Get config
				controlPlane, err := ns.GetControlPlane()
				util.Check(err)
				baseController, err := controlPlane.GetController(name)
				util.Check(err)
				cliCommand := []string{"iofog-controller"}
				switch controller := baseController.(type) {
				case *rsc.KubernetesController:
					k8sControlPlane, ok := controlPlane.(*rsc.KubernetesControlPlane)
					if !ok {
						util.Check(util.NewError("Could not convert Control Plane to Kubernetes Control Plane"))
					}
					util.Check(k8sControlPlane.ValidateKubeConfig())
					k8sExecute(k8sControlPlane.KubeConfig, namespace, "name=controller", cliCommand, args[2:])
				case *rsc.RemoteController:
					if controller.ValidateSSH() != nil {
						util.Check(fmt.Errorf(sshErrMsg, "Controller", controller.Name))
					}
					remoteExec(controller.SSH.User, controller.Host, controller.SSH.KeyFile, controller.SSH.Port, "sudo iofog-controller", args[2:])
				case *rsc.LocalController:
					localExecute(install.GetLocalContainerName("controller", false), cliCommand, args[2:])
				}
			case "agent":
				// Update local cache based on Controller
				err := clientutil.SyncAgentInfo(namespace)
				util.Check(err)
				// Get config
				var baseAgent rsc.Agent
				if useDetached {
					baseAgent, err = config.GetDetachedAgent(name)
				} else {
					baseAgent, err = ns.GetAgent(name)
				}
				util.Check(err)
				switch agent := baseAgent.(type) {
				case *rsc.LocalAgent:
					localExecute(install.GetLocalContainerName("agent", false), []string{"iofog-agent"}, args[2:])
					return
				case *rsc.RemoteAgent:
					// SSH connect
					if agent.ValidateSSH() != nil {
						util.Check(fmt.Errorf(sshErrMsg, "Agent", agent.Name))
					}
					remoteExec(agent.SSH.User, agent.Host, agent.SSH.KeyFile, agent.SSH.Port, "sudo iofog-agent", args[2:])
				}
			default:
				util.Check(util.NewInputError("Unknown legacy CLI " + resource))
			}
		},
	}

	cmd.Flags().Bool("detached", false, pkg.flagDescDetached)

	return cmd
}
