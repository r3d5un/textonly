package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "toctl",
	Short: "CLI tools for managing textonly",
	Long: `textonly-control (toctl) is a collection of tools for managing
the textonly blog application.

The standard textonly web application does not include any UI for
administering the application. Instead, toctl is meant to be the primary
way to manage publish and manage posts.`,
	Version: "0.1",
}

func init() {
	versionTemplate := `{{printf "%s: %s - version %s\n" .Name .Short .Version}}`
	rootCmd.SetVersionTemplate(versionTemplate)
}