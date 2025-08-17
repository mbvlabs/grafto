package ddl

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/mbvlabs/grafto/cmd/generate/catalog"
)

type StatementType int

const (
	CreateTable StatementType = iota
	AlterTable
	DropTable
	CreateIndex
	DropIndex
	CreateSchema
	DropSchema
	CreateEnum
	DropEnum
	Unknown
)

type DDLStatement struct {
	Type        StatementType
	SchemaName  string
	TableName   string
	ColumnDef   *catalog.Column
	Operation   string // ADD, DROP, MODIFY, etc.
	Raw         string // original SQL
	IndexDef    *catalog.Index
	EnumDef     *catalog.Enum
	IfNotExists bool // for CREATE TABLE IF NOT EXISTS
}

func ParseDDLStatement(sql, migrationFile string) (*DDLStatement, error) {
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return nil, nil
	}

	stmt := &DDLStatement{
		Raw: sql,
	}

	sqlLower := strings.ToLower(sql)

	switch {
	case strings.HasPrefix(sqlLower, "create table"):
		return parseCreateTable(sql, migrationFile)
	case strings.HasPrefix(sqlLower, "alter table"):
		return parseAlterTable(sql, migrationFile)
	case strings.HasPrefix(sqlLower, "drop table"):
		return parseDropTable(sql, migrationFile)
	case strings.HasPrefix(sqlLower, "create index") || strings.HasPrefix(sqlLower, "create unique index"):
		return parseCreateIndex(sql, migrationFile)
	case strings.HasPrefix(sqlLower, "drop index"):
		return parseDropIndex(sql, migrationFile)
	case strings.HasPrefix(sqlLower, "create schema"):
		return parseCreateSchema(sql, migrationFile)
	case strings.HasPrefix(sqlLower, "drop schema"):
		return parseDropSchema(sql, migrationFile)
	case strings.HasPrefix(sqlLower, "create type"):
		return parseCreateEnum(sql, migrationFile)
	case strings.HasPrefix(sqlLower, "drop type"):
		return parseDropEnum(sql, migrationFile)
	default:
		stmt.Type = Unknown
		return stmt, nil
	}
}

func parseCreateTable(sql, migrationFile string) (*DDLStatement, error) {
	// Regex to match CREATE TABLE statement with multiline support
	createTableRegex := regexp.MustCompile(`(?is)create\s+table(\s+if\s+not\s+exists)?\s+(?:(\w+)\.)?(\w+)\s*\(\s*(.*?)\s*\)`)

	matches := createTableRegex.FindStringSubmatch(sql)
	if len(matches) < 5 {
		return nil, fmt.Errorf("invalid CREATE TABLE syntax: %s", sql)
	}

	ifNotExists := matches[1] != ""
	schemaName := matches[2]
	tableName := matches[3]
	columnDefs := matches[4]

	table := catalog.NewTable(schemaName, tableName).SetCreatedBy(migrationFile)

	// Parse column definitions
	columns, err := parseColumnDefinitions(columnDefs, migrationFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse column definitions: %w", err)
	}

	for _, col := range columns {
		if err := table.AddColumn(col); err != nil {
			return nil, fmt.Errorf("failed to add column %s: %w", col.Name, err)
		}
	}

	return &DDLStatement{
		Type:        CreateTable,
		SchemaName:  schemaName,
		TableName:   tableName,
		Raw:         sql,
		IfNotExists: ifNotExists,
	}, nil
}

func parseColumnDefinitions(columnDefs, migrationFile string) ([]*catalog.Column, error) {
	var columns []*catalog.Column
	var primaryKeyColumns []string

	// Split column definitions by comma, but handle parentheses
	defs := splitColumnDefinitions(columnDefs)

	for _, def := range defs {
		def = strings.TrimSpace(def)
		if def == "" {
			continue
		}

		defLower := strings.ToLower(def)

		// Handle separate primary key constraint
		if strings.HasPrefix(defLower, "primary key") {
			pkRegex := regexp.MustCompile(`(?i)primary\s+key\s*\(\s*([^)]+)\s*\)`)
			if matches := pkRegex.FindStringSubmatch(def); len(matches) > 1 {
				pkCols := strings.Split(matches[1], ",")
				for _, col := range pkCols {
					primaryKeyColumns = append(primaryKeyColumns, strings.TrimSpace(col))
				}
			}
			continue
		}

		// Skip other constraints for now
		if strings.HasPrefix(defLower, "foreign key") ||
			strings.HasPrefix(defLower, "constraint") ||
			strings.HasPrefix(defLower, "unique") ||
			strings.HasPrefix(defLower, "check") {
			continue
		}

		// Parse individual column definition
		col, err := parseColumnDefinition(def, migrationFile)
		if err != nil {
			return nil, fmt.Errorf("failed to parse column definition '%s': %w", def, err)
		}

		if col != nil {
			columns = append(columns, col)
		}
	}

	// Mark primary key columns
	for _, col := range columns {
		for _, pkCol := range primaryKeyColumns {
			if col.Name == pkCol {
				col.SetPrimaryKey()
			}
		}
	}

	return columns, nil
}

func parseColumnDefinition(def, migrationFile string) (*catalog.Column, error) {
	parts := strings.Fields(def)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid column definition: %s", def)
	}

	columnName := parts[0]
	columnType := parts[1]

	// Parse type with length/precision
	dataType, length, precision, scale := parseDataType(columnType)

	col := catalog.NewColumn(columnName, dataType).SetCreatedBy(migrationFile)

	if length != nil {
		col.SetLength(*length)
	}

	if precision != nil && scale != nil {
		col.SetPrecisionScale(*precision, *scale)
	}

	defLower := strings.ToLower(def)

	// Parse constraints
	if strings.Contains(defLower, "not null") {
		col.SetNotNull()
	}

	if strings.Contains(defLower, "primary key") {
		col.SetPrimaryKey()
	}

	if strings.Contains(defLower, "unique") {
		col.SetUnique()
	}

	// Parse default value
	defaultRegex := regexp.MustCompile(`(?i)default\s+([^,\s]+(?:\s+[^,\s]+)*)`)
	if matches := defaultRegex.FindStringSubmatch(def); len(matches) > 1 {
		col.SetDefault(strings.TrimSpace(matches[1]))
	}

	return col, nil
}

func parseDataType(typeStr string) (dataType string, length *int32, precision *int32, scale *int32) {
	// Handle types with parameters: varchar(255), decimal(10,2), etc.
	typeRegex := regexp.MustCompile(`^(\w+)(?:\(([^)]+)\))?$`)
	matches := typeRegex.FindStringSubmatch(typeStr)

	if len(matches) < 2 {
		return typeStr, nil, nil, nil
	}

	dataType = strings.ToLower(matches[1])

	if len(matches) > 2 && matches[2] != "" {
		params := strings.Split(matches[2], ",")

		if len(params) == 1 {
			// Single parameter (length)
			if val, err := strconv.ParseInt(strings.TrimSpace(params[0]), 10, 32); err == nil {
				length32 := int32(val)
				length = &length32
			}
		} else if len(params) == 2 {
			// Two parameters (precision, scale)
			if val, err := strconv.ParseInt(strings.TrimSpace(params[0]), 10, 32); err == nil {
				precision32 := int32(val)
				precision = &precision32
			}
			if val, err := strconv.ParseInt(strings.TrimSpace(params[1]), 10, 32); err == nil {
				scale32 := int32(val)
				scale = &scale32
			}
		}
	}

	return dataType, length, precision, scale
}

func splitColumnDefinitions(defs string) []string {
	var result []string
	var current strings.Builder
	parenLevel := 0

	for _, char := range defs {
		switch char {
		case '(':
			parenLevel++
			current.WriteRune(char)
		case ')':
			parenLevel--
			current.WriteRune(char)
		case ',':
			if parenLevel == 0 {
				result = append(result, current.String())
				current.Reset()
			} else {
				current.WriteRune(char)
			}
		default:
			current.WriteRune(char)
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}

func parseAlterTable(sql, migrationFile string) (*DDLStatement, error) {
	// Basic ALTER TABLE parsing - will be expanded in Phase 2
	alterRegex := regexp.MustCompile(`(?i)alter\s+table\s+(?:(\w+)\.)?(\w+)\s+(.+)`)
	matches := alterRegex.FindStringSubmatch(sql)

	if len(matches) < 4 {
		return nil, fmt.Errorf("invalid ALTER TABLE syntax: %s", sql)
	}

	schemaName := matches[1]
	tableName := matches[2]
	operation := strings.TrimSpace(matches[3])

	return &DDLStatement{
		Type:       AlterTable,
		SchemaName: schemaName,
		TableName:  tableName,
		Operation:  operation,
		Raw:        sql,
	}, nil
}

func parseDropTable(sql, migrationFile string) (*DDLStatement, error) {
	dropRegex := regexp.MustCompile(`(?i)drop\s+table(?:\s+if\s+exists)?\s+(?:(\w+)\.)?(\w+)`)
	matches := dropRegex.FindStringSubmatch(sql)

	if len(matches) < 3 {
		return nil, fmt.Errorf("invalid DROP TABLE syntax: %s", sql)
	}

	schemaName := matches[1]
	tableName := matches[2]

	return &DDLStatement{
		Type:       DropTable,
		SchemaName: schemaName,
		TableName:  tableName,
		Raw:        sql,
	}, nil
}

func parseCreateIndex(sql, migrationFile string) (*DDLStatement, error) {
	// Placeholder for CREATE INDEX parsing
	return &DDLStatement{
		Type: CreateIndex,
		Raw:  sql,
	}, nil
}

func parseDropIndex(sql, migrationFile string) (*DDLStatement, error) {
	// Placeholder for DROP INDEX parsing
	return &DDLStatement{
		Type: DropIndex,
		Raw:  sql,
	}, nil
}

func parseCreateSchema(sql, migrationFile string) (*DDLStatement, error) {
	// Placeholder for CREATE SCHEMA parsing
	return &DDLStatement{
		Type: CreateSchema,
		Raw:  sql,
	}, nil
}

func parseDropSchema(sql, migrationFile string) (*DDLStatement, error) {
	// Placeholder for DROP SCHEMA parsing
	return &DDLStatement{
		Type: DropSchema,
		Raw:  sql,
	}, nil
}

func parseCreateEnum(sql, migrationFile string) (*DDLStatement, error) {
	// Placeholder for CREATE TYPE (enum) parsing
	return &DDLStatement{
		Type: CreateEnum,
		Raw:  sql,
	}, nil
}

func parseDropEnum(sql, migrationFile string) (*DDLStatement, error) {
	// Placeholder for DROP TYPE parsing
	return &DDLStatement{
		Type: DropEnum,
		Raw:  sql,
	}, nil
}
