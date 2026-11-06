package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"rewind/internal/server"
)

func Execute(version string) error {
	root := &cobra.Command{
		Use:           "rewind",
		Short:         "Rewind backup service",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
	root.AddCommand(newServeCommand(), newVersionCommand(version))
	return root.Execute()
}

func newServeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the server",
		Args:  cobra.NoArgs,
	}
	listen := cmd.Flags().String("listen", "0.0.0.0:8080", "listen address")
	dataDir := cmd.Flags().String("data-dir", "data", "data directory")
	cmd.RunE = func(*cobra.Command, []string) error {
		return server.Run(*listen, *dataDir)
	}
	return cmd
}

func newVersionCommand(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Fprintln(cmd.OutOrStdout(), version)
		},
	}
}
