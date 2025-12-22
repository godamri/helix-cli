package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/godamri/helix-cli/internal/template"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Generate new components (entity, event, handler) in an existing project",
	Long:  `The 'new' command is used for scaffolding new domain entities and components.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var newEntityCmd = &cobra.Command{
	Use:   "entity [name]",
	Short: "Generate a new Domain Entity",
	Long: `Generates boilerplate files for a new domain entity (Schema, Repository, Service).
Note: You must manually wire the dependencies in main.go (Explicit > Implicit).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

		rawName := args[0]
		entityNameTitle := kebabToPascal(rawName)
		entityNameCamel := kebabToCamel(rawName)
		entityNameLower := strings.ToLower(strings.ReplaceAll(rawName, "-", ""))
		entityPluralNameLower := fmt.Sprintf("%ss", entityNameLower)
		entityFileName := strings.ReplaceAll(rawName, "-", "_")

		var driver string
		promptDriver := &survey.Select{
			Message: "Which driver should this entity use?",
			Options: []string{"ent", "pgx"},
			Default: "ent",
			Help:    "Select 'ent' for standard ORM or 'pgx' for raw SQL repository.",
		}
		if err := survey.AskOne(promptDriver, &driver); err != nil {
			return err
		}

		wd, _ := os.Getwd()
		moduleName := getGoModuleName(wd)

		data := template.TemplateData{
			GoModuleName:      moduleName,
			EntityName:        entityNameTitle,
			EntityNameCamel:   entityNameCamel,
			EntityNameLower:   entityNameLower,
			EntityPluralLower: entityPluralNameLower,
			Driver:            driver, // Pass choice
		}

		if TemplateFS == nil {
			return fmt.Errorf("embedded template FS is nil")
		}
		fetcher := template.NewSmartFetcher(TemplateFS, logger)
		gen := template.NewGenerator(data, fetcher)

		files := map[string]string{
			"templates/entity/entity.go.tmpl": filepath.Join(wd, "internal", "core", "entity", fmt.Sprintf("%s.go", entityFileName)),
			// DTO v1
			"templates/entity/dto.go.tmpl":             filepath.Join(wd, "internal", "core", "dto", "v1", fmt.Sprintf("%s.go", entityFileName)),
			"templates/entity/port_service.go.tmpl":    filepath.Join(wd, "internal", "core", "port", fmt.Sprintf("%s_service.go", entityFileName)),
			"templates/entity/port_repository.go.tmpl": filepath.Join(wd, "internal", "core", "port", fmt.Sprintf("%s_repository.go", entityFileName)),
			"templates/entity/service_impl.go.tmpl":    filepath.Join(wd, "internal", "core", "service", fmt.Sprintf("%s_service.go", entityFileName)),
			"templates/entity/repo_impl.go.tmpl":       filepath.Join(wd, "internal", "adapter", "repository", fmt.Sprintf("%s_repository.go", entityFileName)),
			// Handler v1
			"templates/entity/handler_impl.go.tmpl": filepath.Join(wd, "internal", "adapter", "handler", "v1", fmt.Sprintf("%s_handler.go", entityFileName)),
			"templates/entity/ent_schema.go.tmpl":   filepath.Join(wd, "ent", "schema", fmt.Sprintf("%s.go", entityFileName)),
		}

		slog.Info("Generating entity files...", "entity", entityNameTitle, "driver", driver)
		for src, dest := range files {
			if err := gen.ProcessFile(src, dest); err != nil {
				return fmt.Errorf("gen failed for %s: %w", dest, err)
			}
		}

		slog.Info("Running go generate & tidy...")
		if err := exec.Command("go", "generate", "./ent/...").Run(); err != nil {
			slog.Warn("go generate failed (check ent schema)", "error", err)
		}
		exec.Command("go", "mod", "tidy").Run()

		printWiringInstructions(entityNameTitle, entityNameCamel, driver)
		return nil
	},
}

func printWiringInstructions(name, camel, driver string) {
	fmt.Println("\nEntity generated successfully (Version: v1)!")
	fmt.Println("ACTION REQUIRED: Wire dependencies in 'cmd/server/main.go'")
	fmt.Println("---------------------------------------------------------")
	fmt.Printf("// Imports\n")
	fmt.Printf("import handlerV1 \".../internal/adapter/handler/v1\"\n\n")

	fmt.Printf("// Repository (%s)\n", strings.ToUpper(driver))
	if driver == "ent" {
		fmt.Printf("repo%s := repository.New%sRepository(entClient)\n", name, name)
	} else {
		fmt.Printf("repo%s := repository.New%sRepository(stdMainDB)\n", name, name)
	}
	fmt.Printf("svc%s := service.New%sService(repo%s, outboxRepo, txManager)\n\n", name, name, name)

	fmt.Printf("// Handler (V1)\n")
	fmt.Printf("h%sV1 := handlerV1.New%sHandler(svc%s)\n", name, name, name)

	fmt.Printf("// Route (V1)\n")
	fmt.Printf("r.Route(\"/v1/%ss\", func(r chi.Router) {\n", camel)
	fmt.Printf("\tr.Post(\"/\", h%sV1.Create)\n", name)
	fmt.Printf("\tr.Get(\"/{id}\", h%sV1.GetByID)\n", name)
	fmt.Printf("})\n")
	fmt.Println("---------------------------------------------------------")
}

func getGoModuleName(wd string) string {
	data, _ := os.ReadFile(filepath.Join(wd, "go.mod"))
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "module ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}
	return "github.com/godamri/unknown"
}
