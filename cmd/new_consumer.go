package cmd

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	helixTemplate "github.com/godamri/helix-cli/internal/template"
	"github.com/spf13/cobra"
)

var newConsumerCmd = &cobra.Command{
	Use:     "consumer [name] [topic]",
	Short:   "Generate a new Kafka Consumer Handler",
	Example: "  helix-cli new consumer UserCreated user.events.created",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

		rawName := args[0]
		topic := args[1]

		consumerName := kebabToPascal(rawName)
		fileName := fmt.Sprintf("consumer_%s.go", strings.ReplaceAll(strings.ToLower(rawName), "-", "_"))

		wd, _ := os.Getwd()
		targetDir := filepath.Join(wd, "internal", "adapter", "worker")
		targetFile := filepath.Join(targetDir, fileName)

		// Check overlap
		if _, err := os.Stat(targetFile); err == nil {
			return fmt.Errorf("consumer file '%s' already exists", fileName)
		}

		// Prepare Data - Match with templates/consumer/consumer.go.tmpl
		data := struct {
			EntityName   string
			Topic        string
			GoModuleName string
		}{
			EntityName:   consumerName,
			Topic:        topic,
			GoModuleName: getGoModuleName(wd),
		}

		// TemplateFS is an interface (fs.FS), so check against nil.
		if TemplateFS == nil {
			return fmt.Errorf("embedded template FS is nil")
		}
		fetcher := helixTemplate.NewSmartFetcher(TemplateFS, logger)

		tmplPath := "templates/consumer/consumer.go.tmpl"
		content, err := fetcher.ReadFile(tmplPath)
		if err != nil {
			return fmt.Errorf("failed to read template '%s': %w", tmplPath, err)
		}

		// Parse & Execute
		t, err := template.New("consumer").Parse(string(content))
		if err != nil {
			return fmt.Errorf("failed to parse template: %w", err)
		}

		var buf bytes.Buffer
		if err := t.Execute(&buf, data); err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}

		// Write File
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		if err := os.WriteFile(targetFile, buf.Bytes(), 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}

		fmt.Printf("Consumer '%s' generated at %s\n", consumerName, targetFile)
		fmt.Println("Don't forget to register it in 'cmd/server/main.go'!")

		return nil
	},
}
