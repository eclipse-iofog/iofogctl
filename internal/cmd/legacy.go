package cmd

import (
	"github.com/eclipse-iofog/cli/internal/config"
	"github.com/eclipse-iofog/cli/pkg/util"
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
		Use:   "legacy resource resource_name command args...",
		Short: "Execute commands using legacy CLI",
		Long:  `Execute commands using legacy CLI`,
		Example: `iofogctl get all
iofogctl legacy controller my_controller_name status
iofogctl legacy connector my_controller_name status
iofogctl legacy agent my_agent_name status`,
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
				util.Check(err)
				// Connect to cluster
				//Execute
				config, err := clientcmd.BuildConfigFromFlags("", ctrl.KubeConfig)
				util.Check(err)
				// Instantiate Kubernetes client
				clientset, err := kubernetes.NewForConfig(config)
				util.Check(err)
				podList, err := clientset.CoreV1().Pods("iofog").List(metav1.ListOptions{LabelSelector: "name=controller"})
				if err != nil {
					return
				}
				podName := podList.Items[0].Name
				kubeArgs := []string{"exec", podName, "-n", "iofog", "--", "node", "/controller/src/main"}
				kubeArgs = append(kubeArgs, args[2:]...)
				out, err := util.Exec("KUBECONFIG="+ctrl.KubeConfig, "kubectl", kubeArgs...)
				util.Check(err)
				println(out.String())
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

				println(out.String())
			case "connector":
				// Get config
				ctrl, err := config.GetController(namespace, name)
				util.Check(err)
				// Connect to cluster
				//Execute
				config, err := clientcmd.BuildConfigFromFlags("", ctrl.KubeConfig)
				util.Check(err)
				// Instantiate Kubernetes client
				clientset, err := kubernetes.NewForConfig(config)
				util.Check(err)
				podList, err := clientset.CoreV1().Pods("iofog").List(metav1.ListOptions{LabelSelector: "name=connector"})
				if err != nil {
					return
				}
				podName := podList.Items[0].Name
				kubeArgs := []string{"exec", podName, "-n", "iofog", "--", "java", "-jar", "/usr/bin/iofog-connectord.jar"}
				kubeArgs = append(kubeArgs, args[2:]...)
				out, err := util.Exec("KUBECONFIG="+ctrl.KubeConfig, "kubectl", kubeArgs...)
				util.Check(err)
				println(out.String())
			default:
				util.Check(util.NewInputError("Unknown legacy CLI " + resource))
			}
		},
	}

	return cmd
}
