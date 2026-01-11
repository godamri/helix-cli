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

// This assumes a 'new cache' command exists to generate Redis helpers
var newCacheCmd = &cobra.Command{
	Use:   "cache [name]",
	Short: "Generate a new Cache/Redis Repository",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

		rawName := args[0]
		structName := kebabToPascal(rawName)
		fileName := fmt.Sprintf("%s_cache.go", strings.ReplaceAll(strings.ToLower(rawName), "-", "_"))

		wd, _ := os.Getwd()
		targetDir := filepath.Join(wd, "internal", "adapter", "cache")
		targetFile := filepath.Join(targetDir, fileName)

		if _, err := os.Stat(targetFile); err == nil {
			return fmt.Errorf("cache file '%s' already exists", fileName)
		}

		data := struct {
			StructName      string
			LowerStructName string
		}{
			StructName:      structName,
			LowerStructName: strings.ToLower(structName),
		}

		// Use TemplateFS & SmartFetcher
		if TemplateFS == nil {
			return fmt.Errorf("embedded template FS is nil")
		}
		fetcher := helixTemplate.NewSmartFetcher(TemplateFS, logger)

		tmplPath := "templates/cache/cache.go.tmpl"
		content, err := fetcher.ReadFile(tmplPath)
		if err != nil {
			return fmt.Errorf("failed to read template '%s': %w", tmplPath, err)
		}

		t, err := template.New("cache").Parse(string(content))
		if err != nil {
			return fmt.Errorf("failed to parse template: %w", err)
		}

		var buf bytes.Buffer
		if err := t.Execute(&buf, data); err != nil {
			return fmt.Errorf("failed to execute template: %w", err)
		}

		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		if err := os.WriteFile(targetFile, buf.Bytes(), 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}

		fmt.Printf("Cache repository '%s' generated at %s\n", structName, targetFile)
		return nil
	},
}
