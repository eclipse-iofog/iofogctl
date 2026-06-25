package cmd

import (
	reconcileagent "github.com/eclipse-iofog/iofogctl/internal/reconcile/agent"
	reconcileservice "github.com/eclipse-iofog/iofogctl/internal/reconcile/service"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newReconcileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reconcile",
		Short: "Retry async platform provisioning for an agent or service",
		Long:  "Enqueue a manual platform reconcile and wait for provisioning to complete.",
	}

	cmd.AddCommand(
		newReconcileAgentCommand(),
		newReconcileServiceCommand(),
	)
	return cmd
}

func newReconcileAgentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "agent NAME",
		Short:   "Reconcile fog router/NATS platform for an agent",
		Example: "iofogctl reconcile agent my-agent",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			exe := reconcileagent.NewExecutor(namespace, name)
			util.Check(exe.Execute())
			util.PrintSuccess("Successfully reconciled agent " + namespace + "/" + name)
		},
	}
	return cmd
}

func newReconcileServiceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "service NAME",
		Short:   "Reconcile service hub provisioning",
		Example: "iofogctl reconcile service my-service",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			exe := reconcileservice.NewExecutor(namespace, name)
			util.Check(exe.Execute())
			util.PrintSuccess("Successfully reconciled service " + namespace + "/" + name)
		},
	}
	return cmd
}
