package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Manage database migrations via Dockerized Atlas",
	Long: `Generates database migrations using the Atlas binary inside your app container.
	
Prerequisites:
  - Docker & Docker Compose must be installed.
  - The shared infrastructure (DB) must be running (make up-infra).`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var diffCmd = &cobra.Command{
	Use:     "diff [name]",
	Short:   "Generate a new migration file (SQL) inside container",
	Example: "  helix-cli migrate diff add_users_table",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		migrationName := args[0]

		// Check Prerequisites
		if _, err := exec.LookPath("docker"); err != nil {
			return fmt.Errorf("'docker' binary not found. This CLI relies on Docker.")
		}

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get cwd: %w", err)
		}

		// Ensure we are at project root
		if _, err := os.Stat(filepath.Join(cwd, "docker-compose.yml")); os.IsNotExist(err) {
			return fmt.Errorf("docker-compose.yml not found. Please run this command from the project root (e.g., inside svc-order/).")
		}

		// Derive Context (Project Name & DB Name)
		projectName := filepath.Base(cwd)

		dbName := deriveDBName(projectName)

		slog.Info("Running Atlas Diff inside Docker...",
			"migration", migrationName,
			"service", projectName,
			"target_db", dbName,
		)

		// Construct the Docker Command
		devURL := fmt.Sprintf("postgres://dev:dev@helix-shared-db-atlas:5432/%s?sslmode=disable", dbName)

		cmdArgs := []string{
			"compose", "run", "--rm",
			projectName, // Service Name in docker-compose.yml
			"atlas", "migrate", "diff", migrationName,
			"--dir", "file://migrations",
			"--to", "ent://ent/schema",
			"--dev-url", devURL,
		}

		// Execute
		// Interactive mode enabled for prompts
		return runCommand("docker", cmdArgs...)
	},
}

func init() {
	// ONLY diff is allowed via CLI.
	migrateCmd.AddCommand(diffCmd)
}

// runCommand helper to execute shell commands with stdout/stderr attached
func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	slog.Debug("Executing", "cmd", name, "args", strings.Join(args, " "))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}
	return nil
}

// deriveDBName makes a best-effort guess at the DB name based on Helix conventions.
// svc-user-profile -> userprofiles
func deriveDBName(projectName string) string {
	raw := strings.TrimPrefix(projectName, "svc-")
	lower := strings.ToLower(strings.ReplaceAll(raw, "-", ""))
	return fmt.Sprintf("%ss", lower)
}
