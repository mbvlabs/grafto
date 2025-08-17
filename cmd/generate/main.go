package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/jinzhu/inflection"
	"github.com/mbvlabs/grafto/cmd/generate/catalog"
	"github.com/mbvlabs/grafto/cmd/generate/config"
	"github.com/mbvlabs/grafto/cmd/generate/ddl"
	"github.com/mbvlabs/grafto/cmd/generate/generator"
	"github.com/mbvlabs/grafto/cmd/generate/migrations"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run cmd/generate/main.go <type> <name>")
		fmt.Println("Example: go run cmd/generate/main.go view User")
		fmt.Println("Example: go run cmd/generate/main.go model Book")
		os.Exit(1)
	}

	generateType := os.Args[1]
	resourceName := os.Args[2]

	switch generateType {
	case "view":
		if err := generateView(resourceName); err != nil {
			fmt.Printf("Error generating view: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated view for %s\n", resourceName)
	case "model":
		if err := generateModel(resourceName); err != nil {
			fmt.Printf("Error generating model: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated model for %s\n", resourceName)
	case "all":
		if err := generateModel(resourceName); err != nil {
			fmt.Printf("Error generating model: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated model for %s\n", resourceName)

		if err := generateView(resourceName); err != nil {
			fmt.Printf("Error generating view: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated view for %s\n", resourceName)
	default:
		fmt.Printf("Unknown generate type: %s\n", generateType)
		os.Exit(1)
	}
}

func generateModel(resourceName string) error {
	pluralName := inflection.Plural(strings.ToLower(resourceName))

	modelPath := filepath.Join("models", strings.ToLower(resourceName)+".go")
	sqlPath := filepath.Join("psql", pluralName+".sql")

	if _, err := os.Stat(modelPath); err == nil {
		return fmt.Errorf("model file %s already exists", modelPath)
	}

	if _, err := os.Stat(sqlPath); err == nil {
		return fmt.Errorf("SQL file %s already exists", sqlPath)
	}

	cfg := config.NewDefaultConfig()
	cfg.TableName = pluralName

	migrationsList, err := migrations.DiscoverMigrations(cfg.MigrationDirs)
	if err != nil {
		return fmt.Errorf("failed to discover migrations: %w", err)
	}

	if len(migrationsList) == 0 {
		return fmt.Errorf("no migration files found in %v", cfg.MigrationDirs)
	}

	cat := catalog.NewCatalog("public")

	for _, migration := range migrationsList {
		for _, stmt := range migration.Statements {
			if isRelevantForTable(stmt, pluralName) {
				if err := ddl.ApplyDDL(cat, stmt, migration.FilePath); err != nil {
					return fmt.Errorf(
						"failed to apply DDL from %s: %w",
						migration.FilePath,
						err,
					)
				}
			}
		}
	}

	table, err := cat.GetTable("", pluralName)
	if err != nil {
		return fmt.Errorf(
			"table '%s' not found in migrations. Check that you have a CREATE TABLE %s statement in your migrations",
			pluralName,
			pluralName,
		)
	}

	if err := generateSQLFile(resourceName, pluralName, table, sqlPath); err != nil {
		return fmt.Errorf("failed to generate SQL file: %w", err)
	}

	cfg.PackageName = "models"
	model, err := generator.GenerateModel(cat, generator.GeneratorConfig{
		TableName:    pluralName,
		ResourceName: resourceName,
		PackageName:  cfg.PackageName,
		DatabaseType: cfg.DatabaseType,
		StructTags:   cfg.StructTags,
		CustomTypes:  cfg.CustomTypes,
	})
	if err != nil {
		return fmt.Errorf("failed to generate model: %w", err)
	}

	templatePath := filepath.Join("cmd", "generate", "templates", "model.tmpl")
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read model template: %w", err)
	}

	modelContent, err := generator.GenerateModelFile(
		model,
		string(templateContent),
	)
	if err != nil {
		return fmt.Errorf("failed to generate model file content: %w", err)
	}

	if err := os.WriteFile(modelPath, []byte(modelContent), 0644); err != nil {
		return fmt.Errorf("failed to write model file: %w", err)
	}

	if err := runSQLCGenerate(); err != nil {
		return fmt.Errorf("failed to run sqlc generate: %w", err)
	}

	return nil
}

func isRelevantForTable(stmt, targetTable string) bool {
	stmtLower := strings.ToLower(stmt)
	targetLower := strings.ToLower(targetTable)

	if strings.Contains(stmtLower, "create table") &&
		strings.Contains(stmtLower, targetLower) {
		createTableRegex := regexp.MustCompile(
			`(?i)create\s+table(?:\s+if\s+not\s+exists)?\s+(?:\w+\.)?(\w+)`,
		)
		matches := createTableRegex.FindStringSubmatch(stmt)
		if len(matches) > 1 && strings.ToLower(matches[1]) == targetLower {
			return true
		}
	}

	if strings.Contains(stmtLower, "alter table") &&
		strings.Contains(stmtLower, targetLower) {
		alterTableRegex := regexp.MustCompile(
			`(?i)alter\s+table\s+(?:if\s+exists\s+)?(?:\w+\.)?(\w+)`,
		)
		matches := alterTableRegex.FindStringSubmatch(stmt)
		if len(matches) > 1 && strings.ToLower(matches[1]) == targetLower {
			return true
		}
	}

	if strings.Contains(stmtLower, "drop table") &&
		strings.Contains(stmtLower, targetLower) {
		dropTableRegex := regexp.MustCompile(
			`(?i)drop\s+table(?:\s+if\s+exists)?\s+(?:\w+\.)?(\w+)`,
		)
		matches := dropTableRegex.FindStringSubmatch(stmt)
		if len(matches) > 1 && strings.ToLower(matches[1]) == targetLower {
			return true
		}
	}

	return false
}

// func getAvailableTableNames(cat *catalog.Catalog) string {
// 	tables, err := cat.ListTables("")
// 	if err != nil {
// 		return "unable to list tables"
// 	}
//
// 	var names []string
// 	for _, table := range tables {
// 		names = append(names, table.Name)
// 	}
//
// 	return strings.Join(names, ", ")
// }

func generateSQLFile(
	resourceName string,
	pluralName string,
	table *catalog.Table,
	sqlPath string,
) error {
	templatePath := filepath.Join(
		"cmd",
		"generate",
		"templates",
		"crud_operations.sql",
	)
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read SQL template: %w", err)
	}

	var insertColumns []string
	var insertPlaceholders []string
	var updateColumns []string

	placeholderIndex := 1

	for _, col := range table.Columns {
		insertColumns = append(insertColumns, col.Name)
		insertPlaceholders = append(
			insertPlaceholders,
			fmt.Sprintf("$%d", placeholderIndex),
		)
		placeholderIndex++
	}

	placeholderIndex = 2
	for _, col := range table.Columns {
		if col.Name != "id" && col.Name != "created_at" {
			updateColumns = append(
				updateColumns,
				fmt.Sprintf("%s=$%d", col.Name, placeholderIndex),
			)
			placeholderIndex++
		}
	}

	data := struct {
		ResourceName       string
		PluralName         string
		InsertColumns      string
		InsertPlaceholders string
		UpdateColumns      string
	}{
		ResourceName:       resourceName,
		PluralName:         pluralName,
		InsertColumns:      strings.Join(insertColumns, ", "),
		InsertPlaceholders: strings.Join(insertPlaceholders, ", "),
		UpdateColumns:      strings.Join(updateColumns, ", "),
	}

	t, err := template.New("sql").Parse(string(templateContent))
	if err != nil {
		return err
	}

	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return err
	}

	return os.WriteFile(sqlPath, []byte(buf.String()), 0644)
}

func runSQLCGenerate() error {
	cmd := exec.Command("just", "generate-db-functions")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf(
			"failed to run 'just generate-db-functions': %w\nOutput: %s",
			err,
			output,
		)
	}
	fmt.Println("Generated database functions with sqlc")
	return nil
}

func runFormat() error {
	cmd := exec.Command("just", "fmt-go")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf(
			"failed to run 'just fmt-go': %w\nOutput: %s",
			err,
			output,
		)
	}
	fmt.Println("Generated database functions with sqlc")
	return nil
}

// Keep the original view generation code unchanged
func generateView(resourceName string) error {
	pluralName := strings.ToLower(resourceName) + "s"
	viewPath := filepath.Join("views", pluralName+".templ")

	if _, err := os.Stat(viewPath); err == nil {
		return fmt.Errorf("file %s already exists", viewPath)
	}

	content, err := generateViewContent(resourceName, pluralName)
	if err != nil {
		return fmt.Errorf("failed to generate content: %w", err)
	}

	if err := os.WriteFile(viewPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

type TemplateData struct {
	ResourceName string
	PluralName   string
}

func generateViewContent(resourceName, pluralName string) (string, error) {
	tmplPath := filepath.Join("cmd", "generate", "templates", "view.templ")
	tmplContent, err := os.ReadFile(tmplPath)
	if err != nil {
		return "", fmt.Errorf("failed to read template file: %w", err)
	}

	tmpl, err := template.New("view").Parse(string(tmplContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	data := TemplateData{
		ResourceName: resourceName,
		PluralName:   pluralName,
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
