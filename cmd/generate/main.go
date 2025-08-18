package main

import (
	"context"
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
		fmt.Println("Example: go run cmd/generate/main.go ctrl User")
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
	case "ctrl":
		if err := generateController(resourceName); err != nil {
			fmt.Printf("Error generating controller: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated controller for %s\n", resourceName)
	case "all":
		if err := generateModel(resourceName); err != nil {
			fmt.Printf("Error generating model: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated model for %s\n", resourceName)

		if err := generateController(resourceName); err != nil {
			fmt.Printf("Error generating controller: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated controller for %s\n", resourceName)
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

	//nolint:gosec //
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

	//nolint:gosec //
	return os.WriteFile(sqlPath, []byte(buf.String()), 0644)
}

func runSQLCGenerate() error {
	cmd := exec.CommandContext(
		context.Background(),
		"just",
		"generate-db-functions",
	)
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

func runCompileTemplates() error {
	cmd := exec.Command("just", "ct")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf(
			"failed to run 'just ct': %w\nOutput: %s",
			err,
			output,
		)
	}
	fmt.Println("Generated view templates with templ")
	return nil
}

// Keep the original view generation code unchanged
func generateView(resourceName string) error {
	pluralName := inflection.Plural(strings.ToLower(resourceName))
	viewPath := filepath.Join("views", pluralName+"_resource.templ")

	if _, err := os.Stat(viewPath); err == nil {
		return fmt.Errorf("file %s already exists", viewPath)
	}

	content, err := generateViewContent(resourceName, pluralName)
	if err != nil {
		return fmt.Errorf("failed to generate content: %w", err)
	}

	//nolint:gosec //
	if err := os.WriteFile(viewPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	if err := runCompileTemplates(); err != nil {
		return fmt.Errorf("failed to compile templates: %w", err)
	}

	return nil
}

type TemplateData struct {
	ResourceName string
	PluralName   string
	Fields       []ViewField
}

type ViewField struct {
	Name            string
	GoType          string
	GoFormType      string
	DisplayName     string
	IsTimestamp     bool
	InputType       string
	StringConverter string
	DBName          string
	CamelCase       string
	IsSystemField   bool
}

func generateViewContent(resourceName, pluralName string) (string, error) {
	cfg := config.NewDefaultConfig()
	cfg.TableName = pluralName

	migrationsList, err := migrations.DiscoverMigrations(cfg.MigrationDirs)
	if err != nil {
		return "", fmt.Errorf("failed to discover migrations: %w", err)
	}

	if len(migrationsList) == 0 {
		return "", fmt.Errorf("no migration files found in %v", cfg.MigrationDirs)
	}

	cat := catalog.NewCatalog("public")

	for _, migration := range migrationsList {
		for _, stmt := range migration.Statements {
			if isRelevantForTable(stmt, pluralName) {
				if err := ddl.ApplyDDL(cat, stmt, migration.FilePath); err != nil {
					return "", fmt.Errorf(
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
		return "", fmt.Errorf(
			"table '%s' not found in migrations. Check that you have a CREATE TABLE %s statement in your migrations",
			pluralName,
			pluralName,
		)
	}

	typeMapper := generator.NewTypeMapper("postgresql")

	fields := []ViewField{}
	for _, col := range table.Columns {
		if col.Name == "id" {
			continue
		}

		field := ViewField{
			Name:          formatFieldName(col.Name),
			DisplayName:   formatDisplayName(col.Name),
			DBName:        col.Name,
			CamelCase:     formatCamelCase(col.Name),
			IsSystemField: col.Name == "created_at" || col.Name == "updated_at",
		}

		goType, _, _, err := typeMapper.MapSQLTypeToGo(col.DataType, col.IsNullable)
		if err != nil {
			goType = "string"
		}

		field.GoType = goType

		switch goType {
		case "time.Time":
			field.IsTimestamp = true
			field.InputType = "date"
			field.StringConverter = "%s.String()"
		case "string":
			field.InputType = "text"
			field.StringConverter = ""
		case "int16":
			field.InputType = "number"
			field.StringConverter = "fmt.Sprintf(\"%d\", %s)"
		case "int32":
			field.InputType = "number"
			field.StringConverter = "fmt.Sprintf(\"%d\", %s)"
		case "int64":
			field.InputType = "number"
			field.StringConverter = "fmt.Sprintf(\"%d\", %s)"
		case "float32":
			field.InputType = "number"
			field.StringConverter = "fmt.Sprintf(\"%f\", %s)"
		case "float64":
			field.InputType = "number"
			field.StringConverter = "fmt.Sprintf(\"%f\", %s)"
		case "bool":
			field.InputType = "checkbox"
			field.StringConverter = "fmt.Sprintf(\"%t\", %s)"
		case "uuid.UUID":
			field.InputType = "text"
			field.StringConverter = "%s.String()"
		case "[]byte":
			field.InputType = "text"
			field.StringConverter = "string(%s)"
		default:
			field.InputType = "text"
			field.StringConverter = ""
		}

		fields = append(fields, field)
	}

	tmplPath := filepath.Join(
		"cmd",
		"generate",
		"templates",
		"resource_view.tmpl",
	)
	tmplContent, err := os.ReadFile(tmplPath)
	if err != nil {
		return "", fmt.Errorf("failed to read template file: %w", err)
	}

	funcMap := template.FuncMap{
		"ToLower": strings.ToLower,
		"StringDisplay": func(field ViewField, resourceName string) string {
			if field.StringConverter == "" {
				return fmt.Sprintf("{ %s.%s }", strings.ToLower(resourceName), field.Name)
			}
			actualFieldRef := strings.ToLower(resourceName) + "." + field.Name
			converter := strings.ReplaceAll(field.StringConverter, "%s", actualFieldRef)
			return fmt.Sprintf("{ %s }", converter)
		},
		"StringTableDisplay": func(field ViewField, resourceName string) string {
			if field.StringConverter == "" {
				return fmt.Sprintf("{ %s.%s }", strings.ToLower(resourceName), field.Name)
			}
			actualFieldRef := strings.ToLower(resourceName) + "." + field.Name
			converter := strings.ReplaceAll(field.StringConverter, "%s", actualFieldRef)
			return fmt.Sprintf("{ %s }", converter)
		},
		"StringValue": func(field ViewField, resourceName string) string {
			if field.StringConverter == "" {
				return fmt.Sprintf("%s.%s", strings.ToLower(resourceName), field.Name)
			}
			actualFieldRef := strings.ToLower(resourceName) + "." + field.Name
			return strings.ReplaceAll(field.StringConverter, "%s", actualFieldRef)
		},
	}

	tmpl, err := template.New("resource_view").
		Funcs(funcMap).
		Parse(string(tmplContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	data := TemplateData{
		ResourceName: resourceName,
		PluralName:   pluralName,
		Fields:       fields,
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

func formatFieldName(dbName string) string {
	parts := strings.Split(dbName, "_")
	for i, part := range parts {
		parts[i] = strings.Title(part)
	}
	return strings.Join(parts, "")
}

func formatDisplayName(dbName string) string {
	parts := strings.Split(dbName, "_")
	for i, part := range parts {
		parts[i] = strings.Title(part)
	}
	return strings.Join(parts, " ")
}

func formatCamelCase(dbName string) string {
	parts := strings.Split(dbName, "_")
	if len(parts) == 0 {
		return dbName
	}
	
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += strings.Title(parts[i])
	}
	return result
}

func generateController(resourceName string) error {
	pluralName := inflection.Plural(strings.ToLower(resourceName))

	modelPath := filepath.Join("models", strings.ToLower(resourceName)+".go")
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return fmt.Errorf(
			"model file %s does not exist. Generate model first",
			modelPath,
		)
	}

	controllerPath := filepath.Join("controllers", pluralName+".go")
	if _, err := os.Stat(controllerPath); err == nil {
		return fmt.Errorf("controller file %s already exists", controllerPath)
	}

	routesPath := filepath.Join("router", "routes", pluralName+".go")
	if _, err := os.Stat(routesPath); err == nil {
		return fmt.Errorf("routes file %s already exists", routesPath)
	}

	viewPath := filepath.Join("views", pluralName+"_resource.templ")
	if _, err := os.Stat(viewPath); os.IsNotExist(err) {
		if err := generateView(resourceName); err != nil {
			return fmt.Errorf("failed to generate view: %w", err)
		}
		fmt.Printf("Generated view for %s\n", resourceName)
	}

	if err := generateControllerFile(resourceName, pluralName, controllerPath); err != nil {
		return fmt.Errorf("failed to generate controller file: %w", err)
	}

	if err := generateRoutesFile(resourceName, pluralName, routesPath); err != nil {
		return fmt.Errorf("failed to generate routes file: %w", err)
	}

	if err := registerController(resourceName, pluralName); err != nil {
		return fmt.Errorf("failed to register controller: %w", err)
	}

	if err := registerRoutes(resourceName, pluralName); err != nil {
		return fmt.Errorf("failed to register routes: %w", err)
	}

	return nil
}

func generateControllerFile(
	resourceName, pluralName, controllerPath string,
) error {
	// Get field information from database schema
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

	typeMapper := generator.NewTypeMapper("postgresql")

	fields := []ViewField{}
	for _, col := range table.Columns {
		if col.Name == "id" {
			continue
		}

		field := ViewField{
			Name:          formatFieldName(col.Name),
			DisplayName:   formatDisplayName(col.Name),
			DBName:        col.Name,
			CamelCase:     formatCamelCase(col.Name),
			IsSystemField: col.Name == "created_at" || col.Name == "updated_at",
		}

		goType, _, _, err := typeMapper.MapSQLTypeToGo(col.DataType, col.IsNullable)
		if err != nil {
			goType = "string"
		}

		field.GoType = goType

		// Map Go types to appropriate form types
		switch goType {
		case "time.Time":
			field.GoFormType = "time.Time"
			field.IsTimestamp = true
		case "int16":
			field.GoFormType = "int16"
		case "int32":
			field.GoFormType = "int32"
		case "int64":
			field.GoFormType = "int64"
		case "float32":
			field.GoFormType = "float32"
		case "float64":
			field.GoFormType = "float64"
		case "bool":
			field.GoFormType = "bool"
		default:
			field.GoFormType = "string"
		}

		fields = append(fields, field)
	}

	templatePath := filepath.Join(
		"cmd",
		"generate",
		"templates",
		"controller.tmpl",
	)
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read controller template: %w", err)
	}

	funcMap := template.FuncMap{
		"ToLower": strings.ToLower,
	}

	tmpl, err := template.New("controller").
		Funcs(funcMap).
		Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("failed to parse controller template: %w", err)
	}

	data := TemplateData{
		ResourceName: resourceName,
		PluralName:   pluralName,
		Fields:       fields,
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute controller template: %w", err)
	}

	//nolint:gosec //
	if err := os.WriteFile(controllerPath, []byte(buf.String()), 0644); err != nil {
		return fmt.Errorf("failed to write controller file: %w", err)
	}

	return nil
}

func generateRoutesFile(resourceName, pluralName, routesPath string) error {
	templatePath := filepath.Join("cmd", "generate", "templates", "routes.tmpl")
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read routes template: %w", err)
	}

	funcMap := template.FuncMap{
		"ToLower": strings.ToLower,
	}

	tmpl, err := template.New("routes").Funcs(funcMap).Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("failed to parse routes template: %w", err)
	}

	data := TemplateData{
		ResourceName: resourceName,
		PluralName:   pluralName,
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute routes template: %w", err)
	}

	//nolint:gosec //
	if err := os.WriteFile(routesPath, []byte(buf.String()), 0644); err != nil {
		return fmt.Errorf("failed to write routes file: %w", err)
	}

	return nil
}

func registerController(resourceName, pluralName string) error {
	controllerFilePath := "controllers/controller.go"

	content, err := os.ReadFile(controllerFilePath)
	if err != nil {
		return fmt.Errorf("failed to read controller.go: %w", err)
	}

	contentStr := string(content)

	structField := fmt.Sprintf("\t%ss %ss", resourceName, resourceName)
	if !strings.Contains(contentStr, structField) {
		lines := strings.Split(contentStr, "\n")
		for i, line := range lines {
			if strings.Contains(line, "type Controllers struct {") {
				for j := i + 1; j < len(lines); j++ {
					if strings.Contains(lines[j], "}") &&
						!strings.Contains(lines[j], "{") {
						lines = append(
							lines[:j],
							append([]string{structField}, lines[j:]...)...)
						break
					}
				}
				break
			}
		}
		contentStr = strings.Join(lines, "\n")
	}

	newFunctionCall := fmt.Sprintf(
		"\t%s := new%ss(db)",
		strings.ToLower(pluralName),
		resourceName,
	)
	returnField := fmt.Sprintf("\t\t%s,", strings.ToLower(pluralName))

	if !strings.Contains(contentStr, newFunctionCall) {
		newPattern := `books := newBooks(db)`
		if !strings.Contains(contentStr, newPattern) {
			assetsPattern := `assets := newAssets()`
			assetsReplacement := assetsPattern + "\n" + newFunctionCall
			contentStr = strings.Replace(
				contentStr,
				assetsPattern,
				assetsReplacement,
				1,
			)
		}
	}

	if !strings.Contains(contentStr, returnField) {
		returnPattern := `return Controllers{`
		lines := strings.Split(contentStr, "\n")
		for i, line := range lines {
			if strings.Contains(line, returnPattern) {
				for j := i + 1; j < len(lines); j++ {
					if strings.Contains(lines[j], "}") &&
						!strings.Contains(lines[j], "{") {
						// Insert before the closing brace
						lines = append(
							lines[:j],
							append([]string{returnField}, lines[j:]...)...)
						break
					}
				}
				break
			}
		}
		contentStr = strings.Join(lines, "\n")
	}

	//nolint:gosec //
	return os.WriteFile(controllerFilePath, []byte(contentStr), 0644)
}

func registerRoutes(resourceName, pluralName string) error {
	routesFilePath := "router/routes/routes.go"

	content, err := os.ReadFile(routesFilePath)
	if err != nil {
		return fmt.Errorf("failed to read routes.go: %w", err)
	}

	contentStr := string(content)

	appendLine := fmt.Sprintf("\tr = append(r, %ss...)", resourceName)

	if !strings.Contains(contentStr, appendLine) {
		pattern := `var AllRoutes = func() []Route {`
		replacement := pattern + "\n\tvar r []Route"

		if !strings.Contains(contentStr, "var r []Route") {
			contentStr = strings.Replace(contentStr, pattern, replacement, 1)
		}

		returnPattern := `return r`
		returnReplacement := appendLine + "\n\n\t" + returnPattern
		contentStr = strings.Replace(
			contentStr,
			"\t"+returnPattern,
			"\t"+returnReplacement,
			1,
		)
	}

	//nolint:gosec //
	return os.WriteFile(routesFilePath, []byte(contentStr), 0644)
}
