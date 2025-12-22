package main

import (
	"embed"
	"log/slog"
	"os"

	"github.com/godamri/helix-cli/cmd"
	"github.com/godamri/helix-cli/internal/template"
)

// Use a wildcard to embed everything in templates/ recursively
// This ensures .tmpl files are included without manual listing
//
//go:embed templates/Makefile templates/Dockerfile.tmpl templates/Dockerfile.migrate.tmpl templates/.air.toml templates/docker-compose.yml templates/docker-compose.infra.yml templates/.env templates/app templates/entity templates/app/go.mod.tmpl templates/buf.gen.yaml templates/api/proto/v1/service.proto templates/.golangci.yml templates/.github/workflows/ci.yml templates/app/scripts/gen-certs.sh
var templateFS embed.FS

func main() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
	slog.SetDefault(logger)

	// INJECT: Pass the embedded FS to the cmd package
	cmd.TemplateFS = templateFS

	// Check connectivity (Phase 3 Feature)
	template.NetworkCheck()

	cmd.Execute()
}
