package ddl

import (
	"fmt"
	"path/filepath"

	"github.com/mbvlabs/grafto/cmd/generate/catalog"
)

// ApplyDDL applies a DDL statement to the catalog
func ApplyDDL(catalog *catalog.Catalog, stmt string, migrationFile string) error {
	ddlStmt, err := ParseDDLStatement(stmt, migrationFile)
	if err != nil {
		return fmt.Errorf("failed to parse DDL statement in %s: %w", filepath.Base(migrationFile), err)
	}

	if ddlStmt == nil {
		return nil // Skip empty statements
	}

	switch ddlStmt.Type {
	case CreateTable:
		return applyCreateTable(catalog, ddlStmt, migrationFile)
	case AlterTable:
		return applyAlterTable(catalog, ddlStmt, migrationFile)
	case DropTable:
		return applyDropTable(catalog, ddlStmt, migrationFile)
	case CreateIndex:
		return applyCreateIndex(catalog, ddlStmt, migrationFile)
	case DropIndex:
		return applyDropIndex(catalog, ddlStmt, migrationFile)
	case Unknown:
		// Silently skip unknown statements for now
		return nil
	case CreateEnum, DropEnum, CreateSchema, DropSchema:
		// Skip these for now - not needed for basic table scaffolding
		return nil
	default:
		return fmt.Errorf("unsupported DDL statement type: %v in %s", ddlStmt.Type, filepath.Base(migrationFile))
	}
}

func applyCreateTable(catalog *catalog.Catalog, stmt *DDLStatement, migrationFile string) error {
	// Re-parse the CREATE TABLE statement to extract table structure
	table, err := parseCreateTableToTable(stmt.Raw, migrationFile)
	if err != nil {
		return fmt.Errorf("failed to parse CREATE TABLE statement: %w", err)
	}

	schemaName := stmt.SchemaName
	if schemaName == "" {
		schemaName = catalog.DefaultSchema
	}

	// Ensure schema exists
	if _, err := catalog.GetSchema(schemaName); err != nil {
		if _, createErr := catalog.CreateSchema(schemaName); createErr != nil {
			return fmt.Errorf("failed to create schema %s: %w", schemaName, createErr)
		}
	}

	// Handle IF NOT EXISTS
	if stmt.IfNotExists {
		if _, err := catalog.GetTable(schemaName, table.Name); err == nil {
			// Table already exists, skip creation
			return nil
		}
	}

	return catalog.AddTable(schemaName, table)
}

func parseCreateTableToTable(sql, migrationFile string) (*catalog.Table, error) {
	ddlStmt, err := ParseDDLStatement(sql, migrationFile)
	if err != nil {
		return nil, err
	}

	if ddlStmt.Type != CreateTable {
		return nil, fmt.Errorf("expected CREATE TABLE statement")
	}

	// Extract table info from parsed statement
	table := catalog.NewTable(ddlStmt.SchemaName, ddlStmt.TableName).SetCreatedBy(migrationFile)

	// Re-parse column definitions from the raw SQL
	columnDefs, err := extractColumnDefinitions(sql)
	if err != nil {
		return nil, fmt.Errorf("failed to extract column definitions: %w", err)
	}

	columns, err := parseColumnDefinitions(columnDefs, migrationFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse column definitions: %w", err)
	}

	for _, col := range columns {
		if err := table.AddColumn(col); err != nil {
			return nil, fmt.Errorf("failed to add column %s: %w", col.Name, err)
		}
	}

	return table, nil
}

func extractColumnDefinitions(sql string) (string, error) {
	// Extract the content between parentheses in CREATE TABLE
	start := -1
	end := -1
	parenLevel := 0

	for i, char := range sql {
		if char == '(' {
			if start == -1 {
				start = i + 1
			}
			parenLevel++
		} else if char == ')' {
			parenLevel--
			if parenLevel == 0 {
				end = i
				break
			}
		}
	}

	if start == -1 || end == -1 {
		return "", fmt.Errorf("could not find column definitions in CREATE TABLE statement")
	}

	return sql[start:end], nil
}

func applyAlterTable(catalog *catalog.Catalog, stmt *DDLStatement, migrationFile string) error {
	// Basic ALTER TABLE support - will be expanded in Phase 2
	// For now, just return success to avoid breaking existing migrations
	return nil
}

func applyDropTable(catalog *catalog.Catalog, stmt *DDLStatement, migrationFile string) error {
	schemaName := stmt.SchemaName
	if schemaName == "" {
		schemaName = catalog.DefaultSchema
	}

	return catalog.DropTable(schemaName, stmt.TableName)
}

func applyCreateIndex(catalog *catalog.Catalog, stmt *DDLStatement, migrationFile string) error {
	// Placeholder for CREATE INDEX support
	return nil
}

func applyDropIndex(catalog *catalog.Catalog, stmt *DDLStatement, migrationFile string) error {
	// Placeholder for DROP INDEX support
	return nil
}
