package cmd

import (
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/AlecAivazis/survey/v2"
	"github.com/godamri/helix-cli/internal/template"
	"github.com/spf13/cobra"
)

func kebabToPascal(s string) string {
	parts := strings.Split(s, "-")
	var sb strings.Builder
	for _, p := range parts {
		if len(p) > 0 {
			r := []rune(p)
			r[0] = unicode.ToUpper(r[0])
			sb.WriteString(string(r))
		}
	}
	return sb.String()
}

func kebabToCamel(s string) string {
	pascal := kebabToPascal(s)
	if len(pascal) == 0 {
		return ""
	}
	r := []rune(pascal)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

var initCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Initialize a new Helix microservice project (Enterprise Grade)",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

		var projectName string
		if len(args) < 1 {
			prompt := &survey.Input{Message: "What is the project name? (e.g. svc-order)"}
			if err := survey.AskOne(prompt, &projectName); err != nil {
				return err
			}
		} else {
			projectName = args[0]
		}

		if !strings.HasPrefix(projectName, "svc-") {
			var confirm bool
			prompt := &survey.Confirm{
				Message: fmt.Sprintf("Project name '%s' doesn't start with 'svc-'. Auto-fix to 'svc-%s'?", projectName, projectName),
				Default: true,
			}
			if err := survey.AskOne(prompt, &confirm); err != nil {
				return err
			}
			if confirm {
				projectName = "svc-" + projectName
			}
		}

		// --- PREVENT RESERVED NAMES ---
		rawEntityName := strings.TrimPrefix(projectName, "svc-")
		forbidden := map[string]bool{
			"ent":      true,
			"entity":   true,
			"internal": true,
			"pkg":      true,
			"app":      true,
			"go":       true,
		}
		if forbidden[rawEntityName] {
			return fmt.Errorf("FATAL: '%s' is a reserved keyword or framework name. Naming your entity '%s' will break code generation. Please use a real domain name (e.g. svc-user, svc-order).", rawEntityName, kebabToPascal(rawEntityName))
		}

		// --- Driver Selection ---
		var driver string
		promptDriver := &survey.Select{
			Message: "Choose Database Driver Strategy:",
			Options: []string{"ent", "pgx"},
			Default: "ent",
			Help:    "Ent: Type-safe ORM (Productivity). PGX: Raw SQL (Performance/Control).",
		}
		if err := survey.AskOne(promptDriver, &driver); err != nil {
			return err
		}

		destinationDir := fmt.Sprintf("./%s", projectName)
		if _, err := os.Stat(destinationDir); err == nil {
			return fmt.Errorf("directory '%s' already exists", destinationDir)
		}

		r := rand.New(rand.NewSource(time.Now().UnixNano()))

		entityNameTitle := kebabToPascal(rawEntityName)
		entityNameCamel := kebabToCamel(rawEntityName)
		entityNameLower := strings.ToLower(strings.ReplaceAll(rawEntityName, "-", ""))
		entityPluralNameLower := fmt.Sprintf("%ss", entityNameLower)
		goModuleName := fmt.Sprintf("github.com/godamri/%s", projectName)
		entityFileName := strings.ReplaceAll(rawEntityName, "-", "_")

		data := template.TemplateData{
			ProjectName:       projectName,
			GoModuleName:      goModuleName,
			AppPort:           30000 + r.Intn(10000),
			GrpcPort:          30000 + r.Intn(10000),
			DBPort:            40000 + r.Intn(10000),
			DBDevPort:         50000 + r.Intn(10000),
			EntityName:        entityNameTitle,
			EntityNameCamel:   entityNameCamel,
			EntityNameLower:   entityNameLower,
			EntityPluralLower: entityPluralNameLower,
			Driver:            driver, // Pass driver choice
		}

		if TemplateFS == nil {
			return fmt.Errorf("embedded template FS is nil")
		}
		fetcher := template.NewSmartFetcher(TemplateFS, logger)

		generator := template.NewGenerator(data, fetcher)

		templateFiles := map[string]string{
			"templates/Makefile":                                         filepath.Join(destinationDir, "Makefile"),
			"templates/Dockerfile.tmpl":                                  filepath.Join(destinationDir, "Dockerfile"),
			"templates/.air.toml":                                        filepath.Join(destinationDir, ".air.toml"),
			"templates/docker-compose.yml":                               filepath.Join(destinationDir, "docker-compose.yml"),
			"templates/docker-compose.infra.yml":                         filepath.Join(destinationDir, "docker-compose.infra.yml"),
			"templates/.env":                                             filepath.Join(destinationDir, ".env"),
			"templates/.golangci.yml":                                    filepath.Join(destinationDir, ".golangci.yml"),
			"templates/buf.gen.yaml":                                     filepath.Join(destinationDir, "buf.gen.yaml"),
			"templates/api/proto/v1/service.proto":                       filepath.Join(destinationDir, "api", "proto", "v1", fmt.Sprintf("%s.proto", entityFileName)),
			"templates/app/cmd/server/main.go.tmpl":                      filepath.Join(destinationDir, "cmd", "server", "main.go"),
			"templates/app/go.mod.tmpl":                                  filepath.Join(destinationDir, "go.mod"),
			"templates/app/internal/pkg/config/config.go.tmpl":           filepath.Join(destinationDir, "internal", "pkg", "config", "config.go"),
			"templates/app/scripts/gen-certs.sh":                         filepath.Join(destinationDir, "scripts", "gen-certs.sh"),
			"templates/app/ent/runtime.go.tmpl":                          filepath.Join(destinationDir, "ent", "runtime.go"),
			"templates/app/ent/entc.go.tmpl":                             filepath.Join(destinationDir, "ent", "entc.go"),
			"templates/app/ent/generate.go.tmpl":                         filepath.Join(destinationDir, "ent", "generate.go"),
			"templates/entity/ent_schema.go.tmpl":                        filepath.Join(destinationDir, "ent", "schema", fmt.Sprintf("%s.go", entityFileName)),
			"templates/app/ent/schema/outbox.go.tmpl":                    filepath.Join(destinationDir, "ent", "schema", "outbox.go"),
			"templates/entity/port_service.go.tmpl":                      filepath.Join(destinationDir, "internal", "core", "port", fmt.Sprintf("%s_service.go", entityFileName)),
			"templates/entity/port_repository.go.tmpl":                   filepath.Join(destinationDir, "internal", "core", "port", fmt.Sprintf("%s_repository.go", entityFileName)),
			"templates/app/internal/core/port/transaction.go.tmpl":       filepath.Join(destinationDir, "internal", "core", "port", "transaction.go"),
			"templates/app/internal/core/port/outbox_repository.go.tmpl": filepath.Join(destinationDir, "internal", "core", "port", "outbox_repository.go"),
			"templates/entity/entity.go.tmpl":                            filepath.Join(destinationDir, "internal", "core", "entity", fmt.Sprintf("%s.go", entityFileName)),

			// DTO -> V1
			"templates/entity/dto.go.tmpl": filepath.Join(destinationDir, "internal", "core", "dto", "v1", fmt.Sprintf("%s.go", entityFileName)),

			"templates/entity/service_impl.go.tmpl":                               filepath.Join(destinationDir, "internal", "core", "service", fmt.Sprintf("%s_service.go", entityFileName)),
			"templates/entity/repo_impl.go.tmpl":                                  filepath.Join(destinationDir, "internal", "adapter", "repository", fmt.Sprintf("%s_repository.go", entityFileName)),
			"templates/app/internal/adapter/repository/transaction.go.tmpl":       filepath.Join(destinationDir, "internal", "adapter", "repository", "transaction.go"),
			"templates/app/internal/adapter/repository/outbox_repository.go.tmpl": filepath.Join(destinationDir, "internal", "adapter", "repository", "outbox_repository.go"),

			// Handlers -> V1
			"templates/entity/handler_impl.go.tmpl":      filepath.Join(destinationDir, "internal", "adapter", "handler", "v1", fmt.Sprintf("%s_handler.go", entityFileName)),
			"templates/entity/grpc_handler_impl.go.tmpl": filepath.Join(destinationDir, "internal", "adapter", "handler", "v1", fmt.Sprintf("%s_grpc_handler.go", entityFileName)),

			"templates/app/internal/core/entity/errors.go.tmpl":    filepath.Join(destinationDir, "internal", "core", "entity", "errors.go"),
			"templates/app/internal/adapter/worker/outbox.go.tmpl": filepath.Join(destinationDir, "internal", "adapter", "worker", "outbox.go"),

			"templates/app/internal/adapter/handler/validation.go.tmpl": filepath.Join(destinationDir, "internal", "adapter", "handler", "v1", "validation.go"),

			"templates/app/internal/pkg/middleware/deprecation.go.tmpl": filepath.Join(destinationDir, "internal", "pkg", "middleware", "deprecation.go"),

			"templates/app/tests/integration/setup_test.go.tmpl": filepath.Join(destinationDir, "tests", "integration", "setup_test.go"),
			"templates/app/migrations/.keep":                     filepath.Join(destinationDir, "migrations", ".keep"),
			"templates/.github/workflows/ci.yml":                 filepath.Join(destinationDir, ".github", "workflows", "ci.yml"),

			"templates/app/internal/core/port/database.go.tmpl": filepath.Join(destinationDir, "internal", "core", "port", "database.go"),
		}

		slog.Info("Starting scaffolding with SmartFetcher...", "driver", driver)

		for sourcePath, destinationPath := range templateFiles {
			if strings.HasSuffix(filepath.Base(sourcePath), ".keep") {
				os.MkdirAll(filepath.Dir(destinationPath), 0755)
				os.WriteFile(destinationPath, []byte{}, 0644)
				continue
			}

			if err := generator.ProcessFile(sourcePath, destinationPath); err != nil {
				os.RemoveAll(destinationDir)
				return fmt.Errorf("TEMPLATE ERROR: %w", err)
			}
		}

		docsDir := filepath.Join(destinationDir, "docs")
		os.MkdirAll(docsDir, 0755)
		os.WriteFile(filepath.Join(docsDir, "docs.go"), []byte("package docs\n"), 0644)

		logger.Info("Running go mod operations...")
		runShellCommand(destinationDir, "go", "mod", "download")
		runShellCommand(destinationDir, "go", "get", "github.com/kelseyhightower/envconfig")
		runShellCommand(destinationDir, "go", "get", "github.com/go-playground/validator/v10")
		runShellCommand(destinationDir, "go", "get", "github.com/prometheus/client_golang/prometheus/promhttp")
		runShellCommand(destinationDir, "go", "get", "github.com/redis/go-redis/v9")
		runShellCommand(destinationDir, "go", "get", "go.opentelemetry.io/otel")
		runShellCommand(destinationDir, "go", "get", "entgo.io/contrib/entoas")
		runShellCommand(destinationDir, "go", "get", "github.com/ogen-go/ogen")
		runShellCommand(destinationDir, "go", "get", "github.com/testcontainers/testcontainers-go")
		runShellCommand(destinationDir, "go", "get", "github.com/testcontainers/testcontainers-go/modules/postgres")
		runShellCommand(destinationDir, "go", "get", "google.golang.org/grpc")
		runShellCommand(destinationDir, "go", "get", "github.com/swaggo/http-swagger")

		logger.Info("Generating Ent code (Required for schema migration in both modes)...")
		runShellCommand(destinationDir, "go", "generate", "./ent/...")
		// We prepare the artifacts so user can run 'make init'

		logger.Info("Finalizing modules...")
		runShellCommand(destinationDir, "go", "mod", "tidy")
		runShellCommand(destinationDir, "git", "init", "-b", "main")

		fmt.Printf("\nProject %s Initialized Successfully using %s driver!\n", projectName, strings.ToUpper(driver))
		return nil
	},
}

func runShellCommand(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cmd '%s %s' failed:\n%s", name, strings.Join(args, " "), stderr.String())
	}
	return nil
}
