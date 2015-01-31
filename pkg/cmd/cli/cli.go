package cli

import (
	"fmt"
	"os"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/client/clientcmd"
	kubecmd "github.com/GoogleCloudPlatform/kubernetes/pkg/kubectl/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/openshift/origin/pkg/cmd/cli/cmd"
)

const longDesc = `
OpenShift Client

The OpenShift client exposes commands for managing your applications, as well as lower level
tools to interact with each component of your system.

At the present time, the CLI wraps many of the upstream Kubernetes commands and works generically
on all resources.  Some commands you can try:

    $ %[1]s get pods

Note: This is an alpha release of OpenShift and will change significantly.  See
    https://github.com/openshift/origin for the latest information on OpenShift.
`

func NewCommandCLI(name string) *cobra.Command {
	// Main command
	cmds := &cobra.Command{
		Use:     name,
		Aliases: []string{"kubectl"},
		Short:   "Client tools for OpenShift",
		Long:    fmt.Sprintf(longDesc, name),
		Run: func(c *cobra.Command, args []string) {
			c.Help()
		},
	}

	// Override global default to https and port 8443
	clientcmd.DefaultCluster.Server = "https://localhost:8443"

	// TODO: there should be two client configs, one for OpenShift, and one for Kubernetes
	clientConfig := defaultClientConfig(cmds.PersistentFlags())
	f := cmd.NewFactory(clientConfig)
	f.BindFlags(cmds.PersistentFlags())
	out := os.Stdout
	// Kubernetes CRUD commands
	cmds.AddCommand(f.NewCmdGet(out))
	cmds.AddCommand(f.NewCmdDescribe(out))
	cmds.AddCommand(
		// Deprecate 'osc apply' with 'osc create' command.
		applyToCreate(f.NewCmdCreate(out)),
	)
	cmds.AddCommand(f.NewCmdUpdate(out))
	cmds.AddCommand(f.NewCmdDelete(out))
	cmds.AddCommand(kubecmd.NewCmdNamespace(out))

	// Kubernetes support commands
	cmds.AddCommand(f.NewCmdLog(out))
	cmds.AddCommand(f.NewCmdProxy(out))

	// Origin commands
	cmds.AddCommand(cmd.NewCmdProcess(f, out))

	// Origin build commands
	cmds.AddCommand(cmd.NewCmdBuildLogs(f, out))
	cmds.AddCommand(cmd.NewCmdStartBuild(f, out))
	cmds.AddCommand(cmd.NewCmdCancelBuild(f, out))

	cmds.AddCommand(cmd.NewCmdRollback(name, "rollback", f, out))

	return cmds
}

// applyToCreate injects the deprecation notice about for 'apply' command into
// 'create' command.
// TODO: Remove this once we get rid of 'apply' in all documentation/etc.
func applyToCreate(dst *cobra.Command) *cobra.Command {
	dst.Aliases = append(dst.Aliases, "apply")
	oldRun := dst.Run
	dst.Run = func(c *cobra.Command, args []string) {
		if len(os.Args) >= 2 && os.Args[2] == "apply" {
			glog.Errorf("DEPRECATED: The 'apply' command is now deprecated, use 'create' instead.")
		}
		oldRun(c, args)
	}
	return dst
}

// Copy of kubectl/cmd/DefaultClientConfig, using NewNonInteractiveDeferredLoadingClientConfig
func defaultClientConfig(flags *pflag.FlagSet) clientcmd.ClientConfig {
	loadingRules := clientcmd.NewClientConfigLoadingRules()
	loadingRules.EnvVarPath = os.Getenv(clientcmd.RecommendedConfigPathEnvVar)
	flags.StringVar(&loadingRules.CommandLinePath, "kubeconfig", "", "Path to the kubeconfig file to use for CLI requests.")

	overrides := &clientcmd.ConfigOverrides{}
	overrideFlags := clientcmd.RecommendedConfigOverrideFlags("")
	overrideFlags.ContextOverrideFlags.NamespaceShort = "n"
	clientcmd.BindOverrideFlags(overrides, flags, overrideFlags)
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)

	return clientConfig
}
