package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

const DefaultTemplateRepo = "https://github.com/godamri/helix-templates.git"

var updateTemplatesCmd = &cobra.Command{
	Use:   "update-templates [repo-url]",
	Short: "Sync local templates with a remote Git repository",
	Long: fmt.Sprintf(`Clones or pulls the specified Git repository into ~/.helix/templates to override built-in templates.

Default Repository: %s`, DefaultTemplateRepo),
	Example: `  helix-cli update-templates (uses default official repo)
  helix-cli update-templates https://github.com/my-org/custom-templates.git (uses custom fork)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Setup Path
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home dir: %w", err)
		}
		localTemplateDir := filepath.Join(home, ".helix", "templates")

		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

		// Check if directory exists
		if _, err := os.Stat(localTemplateDir); os.IsNotExist(err) {
			// --- FRESH INSTALL ---
			repoURL := DefaultTemplateRepo
			if len(args) > 0 {
				repoURL = args[0]
			}

			logger.Info("Local templates not found. Cloning...", "repo", repoURL, "dest", localTemplateDir)

			// Ensure parent dir exists (.helix)
			if err := os.MkdirAll(filepath.Dir(localTemplateDir), 0755); err != nil {
				return fmt.Errorf("failed to create config dir: %w", err)
			}

			c := exec.Command("git", "clone", repoURL, localTemplateDir)
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if err := c.Run(); err != nil {
				return fmt.Errorf("git clone failed: %w", err)
			}
			fmt.Println("Templates installed successfully.")

		} else {
			// --- UPDATE ---
			logger.Info("Updating templates...", "path", localTemplateDir)

			// We trust 'git pull' to use the origin configured in .git/config
			c := exec.Command("git", "pull")
			c.Dir = localTemplateDir
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if err := c.Run(); err != nil {
				// Fallback diagnostic
				if _, err := os.Stat(filepath.Join(localTemplateDir, ".git")); os.IsNotExist(err) {
					return fmt.Errorf("directory exists but is not a git repository. Remove it manually: rm -rf %s", localTemplateDir)
				}
				return fmt.Errorf("git pull failed: %w", err)
			}
			fmt.Println("Templates updated successfully.")
		}

		return nil
	},
}
