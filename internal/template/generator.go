package template

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

type TemplateData struct {
	ProjectName       string
	GoModuleName      string
	AppPort           int
	GrpcPort          int
	DBPort            int
	DBDevPort         int
	EntityName        string
	EntityNameCamel   string
	EntityNameLower   string
	EntityPluralLower string
	Driver            string
}

type Generator struct {
	Data    TemplateData
	Fetcher Fetcher
}

func NewGenerator(data TemplateData, fetcher Fetcher) *Generator {
	return &Generator{
		Data:    data,
		Fetcher: fetcher,
	}
}

// ProcessFile reads a template from the fetcher, executes it with Data, and writes to dest.
func (g *Generator) ProcessFile(sourcePath, destPath string) error {
	// Read Content via Fetcher
	content, err := g.Fetcher.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("read template '%s': %w", sourcePath, err)
	}

	// Parse Template
	tmpl, err := template.New(filepath.Base(sourcePath)).Parse(string(content))
	if err != nil {
		return fmt.Errorf("parse template '%s': %w", sourcePath, err)
	}

	// Execute
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, g.Data); err != nil {
		return fmt.Errorf("execute template '%s': %w", sourcePath, err)
	}

	// Write to Disk
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	return os.WriteFile(destPath, buf.Bytes(), 0644)
}
