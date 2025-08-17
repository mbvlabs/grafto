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
		if err := generateModelWithNewArchitecture(resourceName); err != nil {
			fmt.Printf("Error generating model: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated model for %s\n", resourceName)
	default:
		fmt.Printf("Unknown generate type: %s\n", generateType)
		os.Exit(1)
	}
}

func generateModelWithNewArchitecture(resourceName string) error {
	pluralName := inflection.Plural(strings.ToLower(resourceName))
	
	modelPath := filepath.Join("models", strings.ToLower(resourceName)+".go")
	sqlPath := filepath.Join("psql", pluralName+".sql")

	// Check if files already exist
	if _, err := os.Stat(modelPath); err == nil {
		return fmt.Errorf("model file %s already exists", modelPath)
	}

	if _, err := os.Stat(sqlPath); err == nil {
		return fmt.Errorf("SQL file %s already exists", sqlPath)
	}

	// Step 1: Discover and parse migrations
	cfg := config.NewDefaultConfig()
	cfg.TableName = pluralName

	migrationsList, err := migrations.DiscoverMigrations(cfg.MigrationDirs)
	if err != nil {
		return fmt.Errorf("failed to discover migrations: %w", err)
	}

	if len(migrationsList) == 0 {
		return fmt.Errorf("no migration files found in %v", cfg.MigrationDirs)
	}

	// Step 2: Build catalog by tracking only the target table through migrations
	cat := catalog.NewCatalog("public")
	
	for _, migration := range migrationsList {
		for _, stmt := range migration.Statements {
			// Only process DDL statements that affect our target table
			if isRelevantForTable(stmt, pluralName) {
				if err := ddl.ApplyDDL(cat, stmt, migration.FilePath); err != nil {
					return fmt.Errorf("failed to apply DDL from %s: %w", migration.FilePath, err)
				}
			}
		}
	}

	// Step 3: Check if target table exists
	table, err := cat.GetTable("", pluralName)
	if err != nil {
		return fmt.Errorf("table '%s' not found in migrations. Check that you have a CREATE TABLE %s statement in your migrations", 
			pluralName, pluralName)
	}

	// Step 4: Generate SQL queries file
	if err := generateSQLFile(resourceName, pluralName, table, sqlPath); err != nil {
		return fmt.Errorf("failed to generate SQL file: %w", err)
	}

	// Step 5: Generate Go model using new architecture
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

	// Generate model file content using template
	modelContent, err := generator.GenerateModelFile(model, generator.GetDefaultModelTemplate())
	if err != nil {
		return fmt.Errorf("failed to generate model file content: %w", err)
	}

	// Write model file
	if err := os.WriteFile(modelPath, []byte(modelContent), 0644); err != nil {
		return fmt.Errorf("failed to write model file: %w", err)
	}

	// Step 6: Run sqlc generate
	if err := runSQLCGenerate(); err != nil {
		return fmt.Errorf("failed to run sqlc generate: %w", err)
	}

	return nil
}

func isRelevantForTable(stmt, targetTable string) bool {
	stmtLower := strings.ToLower(stmt)
	targetLower := strings.ToLower(targetTable)
	
	// Check if this is a CREATE TABLE statement for our target
	if strings.Contains(stmtLower, "create table") && strings.Contains(stmtLower, targetLower) {
		// More precise check to avoid false positives
		createTableRegex := regexp.MustCompile(`(?i)create\s+table(?:\s+if\s+not\s+exists)?\s+(?:\w+\.)?(\w+)`)
		matches := createTableRegex.FindStringSubmatch(stmt)
		if len(matches) > 1 && strings.ToLower(matches[1]) == targetLower {
			return true
		}
	}
	
	// Check if this is an ALTER TABLE statement for our target
	if strings.Contains(stmtLower, "alter table") && strings.Contains(stmtLower, targetLower) {
		alterTableRegex := regexp.MustCompile(`(?i)alter\s+table\s+(?:\w+\.)?(\w+)`)
		matches := alterTableRegex.FindStringSubmatch(stmt)
		if len(matches) > 1 && strings.ToLower(matches[1]) == targetLower {
			return true
		}
	}
	
	// Check if this is a DROP TABLE statement for our target
	if strings.Contains(stmtLower, "drop table") && strings.Contains(stmtLower, targetLower) {
		dropTableRegex := regexp.MustCompile(`(?i)drop\s+table(?:\s+if\s+exists)?\s+(?:\w+\.)?(\w+)`)
		matches := dropTableRegex.FindStringSubmatch(stmt)
		if len(matches) > 1 && strings.ToLower(matches[1]) == targetLower {
			return true
		}
	}
	
	return false
}

func getAvailableTableNames(cat *catalog.Catalog) string {
	tables, err := cat.ListTables("")
	if err != nil {
		return "unable to list tables"
	}
	
	var names []string
	for _, table := range tables {
		names = append(names, table.Name)
	}
	
	return strings.Join(names, ", ")
}

func generateSQLFile(resourceName, pluralName string, table *catalog.Table, sqlPath string) error {
	tmpl := `-- name: Query{{.ResourceName}}ByID :one
select * from {{.PluralName}} where id=$1;

-- name: Query{{.ResourceName}}s :many
select * from {{.PluralName}};

-- name: Insert{{.ResourceName}} :one
insert into
    {{.PluralName}} ({{.InsertColumns}})
values
    ({{.InsertPlaceholders}})
returning *;

-- name: Update{{.ResourceName}} :one
update {{.PluralName}}
    set {{.UpdateColumns}}
where id = $1
returning *;

-- name: Delete{{.ResourceName}} :exec
delete from {{.PluralName}} where id=$1;

-- name: QueryPaginated{{.ResourceName}}s :many
select * from {{.PluralName}} 
order by created_at desc 
limit sqlc.arg('limit')::bigint offset sqlc.arg('offset')::bigint;

-- name: Count{{.ResourceName}}s :one
select count(*) from {{.PluralName}};
`

	var insertColumns []string
	var insertPlaceholders []string
	var updateColumns []string
	
	placeholderIndex := 1
	
	for _, col := range table.Columns {
		if col.Name != "id" {
			insertColumns = append(insertColumns, col.Name)
			insertPlaceholders = append(insertPlaceholders, fmt.Sprintf("$%d", placeholderIndex))
			placeholderIndex++
		}
	}
	
	placeholderIndex = 2 // Start from 2 since $1 is the ID in update
	for _, col := range table.Columns {
		if col.Name != "id" && col.Name != "created_at" {
			updateColumns = append(updateColumns, fmt.Sprintf("%s=$%d", col.Name, placeholderIndex))
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

	t, err := template.New("sql").Parse(tmpl)
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
		return fmt.Errorf("failed to run 'just generate-db-functions': %w\nOutput: %s", err, output)
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