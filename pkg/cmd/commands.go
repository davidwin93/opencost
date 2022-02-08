package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/kubecost/cost-model/pkg/cmd/agent"
	"github.com/kubecost/cost-model/pkg/cmd/costmodel"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog"
)

const (
	// commandRoot is the root command used to route to sub-commands
	commandRoot string = "root"

	// CommandCostModel is the command used to execute the metrics emission and ETL pipeline
	CommandCostModel string = "cost-model"

	// CommandAgent executes the application in agent mode, which provides only metrics exporting.
	CommandAgent string = "agent"
)

// Execute runs the root command for the application. By default, if no command argument is provided,
// on the command line, the cost-model is executed by default.
//
// This function accepts a costModelCmd parameter to provide support for various cost-model implementations
// (ie: open source, enterprise).
func Execute(costModelCmd *cobra.Command) error {
	// use the open-source cost-model if a command is not provided
	if costModelCmd == nil {
		costModelCmd = newCostModelCommand()
	}

	// validate the command being passed
	if err := validate(costModelCmd); err != nil {
		return err
	}

	rootCmd := newRootCommand(costModelCmd)

	// initialize klog and make cobra aware of all the go flags
	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlag(flag.CommandLine.Lookup("v"))
	pflag.CommandLine.AddGoFlag(flag.CommandLine.Lookup("logtostderr"))
	pflag.CommandLine.Set("v", "3")

	// in the event that no directive/command is passed, we want to default to using the cost-model command
	// cobra doesn't provide a way within the API to do this, so we'll prepend the command if it is omitted.
	if len(os.Args) > 1 {
		// try to find the sub-command from the arguments, if there's an error or the command _is_ the
		// root command, prepend the default command
		pCmd, _, err := rootCmd.Find(os.Args[1:])
		if err != nil || pCmd.Use == rootCmd.Use {
			rootCmd.SetArgs(append([]string{CommandCostModel}, os.Args[1:]...))
		}
	} else {
		rootCmd.SetArgs([]string{CommandCostModel})
	}

	return rootCmd.Execute()
}

// newRootCommand creates a new root command which will act as a sub-command router for the
// cost-model application
func newRootCommand(costModelCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          commandRoot,
		SilenceUsage: true,
	}

	// add the modes of operation
	cmd.AddCommand(
		costModelCmd,
		newAgentCommand(),
	)

	return cmd
}

// default open-source cost-model command
func newCostModelCommand() *cobra.Command {
	opts := &costmodel.CostModelOpts{}

	cmCmd := &cobra.Command{
		Use:   CommandCostModel,
		Short: "Cost-Model metric exporter and data aggregator.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return costmodel.Execute(opts)
		},
	}

	// TODO: We could introduce a way of mapping input command-line parameters to a configuration
	// TODO: object, and pass that object to the agent Execute()
	// cmCmd.Flags().<Type>VarP(&opts.<Property>, "<flag>", "<short>", <default>, "<usage>")

	return cmCmd
}

func newAgentCommand() *cobra.Command {
	opts := &agent.AgentOpts{}

	agentCmd := &cobra.Command{
		Use:   CommandAgent,
		Short: "Agent mode operates as a metric exporter only.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return agent.Execute(opts)
		},
	}

	// TODO: We could introduce a way of mapping input command-line parameters to a configuration
	// TODO: object, and pass that object to the agent Execute()
	// agentCmd.Flags().<Type>VarP(&opts.<Property>, "<flag>", "<short>", <default>, "<usage>")

	return agentCmd
}

// validate will check to ensure that the cost model command passed in has a use equal to the
// CommandCostModel to ensure that the default command matches.
func validate(costModelCommand *cobra.Command) error {
	if costModelCommand.Use != CommandCostModel {
		return fmt.Errorf("Incompatible 'cost-model' command provided to run-time.")
	}
	return nil
}
