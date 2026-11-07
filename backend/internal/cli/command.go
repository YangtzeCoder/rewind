package cli

import (
	"fmt"
	"os"
	"strings"

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
	listen := cmd.Flags().String("listen", envOr("REWIND_LISTEN", "0.0.0.0:8080"), "listen address")
	dataDir := cmd.Flags().String("data-dir", envOr("REWIND_DATA_DIR", "data"), "data directory")
	cmd.RunE = func(*cobra.Command, []string) error {
		return server.Run(*listen, *dataDir, initialAdminPassword)
	}
	return cmd
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func initialAdminPassword() (string, error) {
	password, passwordSet := os.LookupEnv("REWIND_ADMIN_PASSWORD")
	passwordFile, passwordFileSet := os.LookupEnv("REWIND_ADMIN_PASSWORD_FILE")
	if passwordSet && passwordFileSet {
		return "", fmt.Errorf("set only one of REWIND_ADMIN_PASSWORD or REWIND_ADMIN_PASSWORD_FILE")
	}
	if passwordFileSet {
		if passwordFile == "" {
			return "", fmt.Errorf("REWIND_ADMIN_PASSWORD_FILE must not be empty")
		}
		contents, err := os.ReadFile(passwordFile)
		if err != nil {
			return "", fmt.Errorf("read initial administrator password file: %w", err)
		}
		password = strings.TrimRight(string(contents), "\r\n")
	}
	return password, nil
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
