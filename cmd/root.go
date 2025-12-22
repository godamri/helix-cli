package cmd

import (
	"io/fs"

	"github.com/spf13/cobra"
)

// Global variable to hold the embedded templates
// Injected from main.go
var TemplateFS fs.FS

var rootCmd = &cobra.Command{
	Use:   "helix-cli",
	Short: "Helix Enterprise Microservice Generator",
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(updateTemplatesCmd)

	var newCmd = &cobra.Command{
		Use:   "new",
		Short: "Generate new components (consumer, entity, cache)",
	}
	newCmd.AddCommand(newEntityCmd)
	newCmd.AddCommand(newConsumerCmd)
	newCmd.AddCommand(newCacheCmd) // Register Cache Command

	rootCmd.AddCommand(newCmd)
}
