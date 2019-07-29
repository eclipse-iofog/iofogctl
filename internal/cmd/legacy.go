/*
 *  *******************************************************************************
 *  * Copyright (c) 2019 Edgeworx, Inc.
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
	"fmt"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"strings"
)

// NOTE: (Serge) This code will be discarded eventually. Keeping it one file.
func newLegacyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "legacy resource RESOURCE COMMAND ARGS...",
		Short: "Execute commands using legacy CLI",
		Long:  `Execute commands using legacy CLI`,
		Example: `iofogctl get all
iofogctl legacy controller NAME iofog
iofogctl legacy connector NAME status
iofogctl legacy agent NAME status`,
		Args: cobra.MinimumNArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			// Get resource type arg
			resource := args[0]
			// Get resource name
			name := args[1]

			// Get namespace option
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			switch resource {
			case "controller":
				// Get config
				ctrl, err := config.GetController(namespace, name)
				if ctrl.KubeConfig != "" {
					util.Check(err)
					ctrl.KubeConfig, err = util.FormatPath(ctrl.KubeConfig)
					util.Check(err)
					// Connect to cluster
					//Execute
					config, err := clientcmd.BuildConfigFromFlags("", ctrl.KubeConfig)
					util.Check(err)
					// Instantiate Kubernetes client
					clientset, err := kubernetes.NewForConfig(config)
					util.Check(err)
					podList, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "name=controller"})
					if err != nil {
						return
					}
					podName := podList.Items[0].Name
					kubeArgs := []string{"exec", podName, "-n", namespace, "--", "iofog-controller"}
					kubeArgs = append(kubeArgs, args[2:]...)
					out, err := util.Exec("KUBECONFIG="+ctrl.KubeConfig, "kubectl", kubeArgs...)
					util.Check(err)
					fmt.Print(out.String())
				} else {
					if ctrl.Host == "" || ctrl.User == "" || ctrl.KeyFile == "" || ctrl.Port == 0 {
						util.PrintNotify(`This client does not have any means of performing legacy commands with the specified Controller.
  This usually means you did not deploy the Controller but instead connected to it after its deployment.
  If it is a Kubernetes-deployed Controller, you can try connecting with the correct Kube Config file.
  If it is a non-Kubernetes-deploy Controller, you must manually add host, user, port, and keyfile fields to ~/.iofog/config.yaml.`)
						util.Check(util.NewError("Could not SSH into Controller to execute legacy command"))
					}
					sshClient := util.NewSecureShellClient(ctrl.User, ctrl.Host, ctrl.KeyFile)
					util.Check(sshClient.Connect())
					defer sshClient.Disconnect()

					sshCmd := "iofog-controller"
					for _, arg := range args[2:] {
						sshCmd = sshCmd + " " + arg
					}
					logs, err := sshClient.Run(sshCmd)
					util.Check(err)
					fmt.Print(logs.String())
				}
			case "agent":
				// Get config
				agent, err := config.GetAgent(namespace, name)
				util.Check(err)
				// SSH connect
				ssh := util.NewSecureShellClient(agent.User, agent.Host, agent.KeyFile)
				util.Check(ssh.Connect())
				// Execute
				args[0] = "sudo"
				args[1] = "iofog-agent"
				command := strings.Join(args, " ")
				out, err := ssh.Run(command)
				util.Check(err)

				fmt.Print(out.String())
			case "connector":
				// Get config
				ctrl, err := config.GetController(namespace, name)
				util.Check(err)
				ctrl.KubeConfig, err = util.FormatPath(ctrl.KubeConfig)
				util.Check(err)
				// Connect to cluster
				//Execute
				config, err := clientcmd.BuildConfigFromFlags("", ctrl.KubeConfig)
				util.Check(err)
				// Instantiate Kubernetes client
				clientset, err := kubernetes.NewForConfig(config)
				util.Check(err)
				podList, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: "name=connector"})
				if err != nil {
					return
				}
				podName := podList.Items[0].Name
				kubeArgs := []string{"exec", podName, "-n", namespace, "--", "iofog-connector"}
				kubeArgs = append(kubeArgs, args[2:]...)
				out, err := util.Exec("KUBECONFIG="+ctrl.KubeConfig, "kubectl", kubeArgs...)
				util.Check(err)
				fmt.Print(out.String())
			default:
				util.Check(util.NewInputError("Unknown legacy CLI " + resource))
			}
		},
	}

	return cmd
}
