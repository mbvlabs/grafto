package generator

import (
	"fmt"
	"strings"

	"github.com/mbvlabs/grafto/cmd/generate/catalog"
)

type TypeMapper struct {
	DatabaseType string
	TypeMap      map[string]string
	Overrides    []TypeOverride
}

type TypeOverride struct {
	DatabaseType string
	GoType       string
	Package      string
	Nullable     bool
}

type GeneratedField struct {
	Name                    string
	Type                    string
	Tag                     string
	Comment                 string
	Package                 string
	SQLCType                string // The type used by sqlc
	ConversionFromDB        string // How to convert from sqlc type to Go type
	ConversionToDB          string // How to convert from Go type to sqlc type
	ConversionToDBForUpdate string // How to convert from Go type to sqlc type in update operations
	ZeroCheck               string // How to check if the value is zero/empty
}

type GeneratedModel struct {
	Name      string
	Package   string
	Fields    []GeneratedField
	Imports   []string
	TableName string
}

func NewTypeMapper(databaseType string) *TypeMapper {
	tm := &TypeMapper{
		DatabaseType: databaseType,
		TypeMap:      make(map[string]string),
		Overrides:    make([]TypeOverride, 0),
	}

	// Initialize default PostgreSQL mappings
	if databaseType == "postgresql" {
		tm.initPostgreSQLMappings()
	}

	return tm
}

func (tm *TypeMapper) initPostgreSQLMappings() {
	// Non-nullable type mappings
	tm.TypeMap["uuid"] = "uuid.UUID"
	tm.TypeMap["varchar"] = "string"
	tm.TypeMap["text"] = "string"
	tm.TypeMap["char"] = "string"
	tm.TypeMap["bytea"] = "[]byte"
	tm.TypeMap["bool"] = "bool"
	tm.TypeMap["boolean"] = "bool"
	tm.TypeMap["timestamp"] = "time.Time"
	tm.TypeMap["timestamptz"] = "time.Time"
	tm.TypeMap["timestamp with time zone"] = "time.Time"
	tm.TypeMap["timestamp without time zone"] = "time.Time"
	tm.TypeMap["date"] = "time.Time"
	tm.TypeMap["time"] = "time.Time"
	tm.TypeMap["jsonb"] = "interface{}"
	tm.TypeMap["json"] = "interface{}"
	tm.TypeMap["int"] = "int32"
	tm.TypeMap["integer"] = "int32"
	tm.TypeMap["int4"] = "int32"
	tm.TypeMap["serial"] = "int32"
	tm.TypeMap["bigint"] = "int64"
	tm.TypeMap["int8"] = "int64"
	tm.TypeMap["bigserial"] = "int64"
	tm.TypeMap["smallint"] = "int16"
	tm.TypeMap["int2"] = "int16"
	tm.TypeMap["smallserial"] = "int16"
	tm.TypeMap["decimal"] = "float64"
	tm.TypeMap["numeric"] = "float64"
	tm.TypeMap["real"] = "float32"
	tm.TypeMap["float4"] = "float32"
	tm.TypeMap["double precision"] = "float64"
	tm.TypeMap["float8"] = "float64"
}

func (tm *TypeMapper) MapSQLTypeToGo(
	sqlType string,
	nullable bool,
) (goType, sqlcType, packageName string, err error) {
	// Normalize the SQL type
	normalizedType := normalizeSQLType(sqlType)

	// Check for overrides first
	for _, override := range tm.Overrides {
		if override.DatabaseType == normalizedType &&
			override.Nullable == nullable {
			return override.GoType, "", override.Package, nil
		}
	}

	// Get base Go type
	baseGoType, exists := tm.TypeMap[normalizedType]
	if !exists {
		return "interface{}", "interface{}", "", nil
	}

	// Determine package requirements
	var pkg string
	switch baseGoType {
	case "uuid.UUID":
		pkg = "github.com/google/uuid"
	case "time.Time":
		pkg = "time"
	}

	// For nullable columns, determine sqlc type and conversion
	if nullable {
		sqlcType, goType = tm.mapNullableType(normalizedType, baseGoType)
	} else {
		goType = baseGoType
		sqlcType = baseGoType
	}

	return goType, sqlcType, pkg, nil
}

func (tm *TypeMapper) mapNullableType(
	sqlType, baseGoType string,
) (sqlcType, goType string) {
	switch baseGoType {
	case "time.Time":
		return "pgtype.Timestamptz", "time.Time"
	case "string":
		return "sql.NullString", "string"
	case "bool":
		return "sql.NullBool", "bool"
	case "int32":
		return "sql.NullInt32", "int32"
	case "int64":
		return "sql.NullInt64", "int64"
	case "float64":
		if sqlType == "decimal" || sqlType == "numeric" {
			return "pgtype.Numeric", "float64"
		}
		return "sql.NullFloat64", "float64"
	case "[]byte":
		return "[]byte", "[]byte" // bytea is typically not nullable in practice
	case "uuid.UUID":
		return "uuid.UUID", "uuid.UUID" // UUIDs are typically not nullable
	default:
		return "interface{}", baseGoType
	}
}

func (tm *TypeMapper) GenerateConversionFromDB(field GeneratedField) string {
	if field.SQLCType == field.Type {
		return fmt.Sprintf("row.%s", field.Name)
	}

	switch field.SQLCType {
	case "pgtype.Timestamptz":
		return fmt.Sprintf("row.%s.Time", field.Name)
	case "sql.NullString":
		return fmt.Sprintf("row.%s.String", field.Name)
	case "sql.NullBool":
		return fmt.Sprintf("row.%s.Bool", field.Name)
	case "sql.NullInt32":
		return fmt.Sprintf("row.%s.Int32", field.Name)
	case "sql.NullInt64":
		return fmt.Sprintf("row.%s.Int64", field.Name)
	case "sql.NullFloat64":
		return fmt.Sprintf("row.%s.Float64", field.Name)
	case "pgtype.Numeric":
		return fmt.Sprintf("func() float64 { if row.%s.Valid { f, _ := row.%s.Float64Value(); return f.Float64 }; return 0 }()", field.Name, field.Name)
	default:
		return fmt.Sprintf("row.%s", field.Name)
	}
}

func (tm *TypeMapper) GenerateConversionToDB(
	field GeneratedField,
	valueExpr string,
) string {
	if field.SQLCType == field.Type {
		return valueExpr
	}

	switch field.SQLCType {
	case "pgtype.Timestamptz", "time.Time":
		return fmt.Sprintf(
			"pgtype.Timestamptz{Time: %s, Valid: true}",
			valueExpr,
		)
	case "sql.NullString":
		return fmt.Sprintf(
			"sql.NullString{String: %s, Valid: %s != \"\"}",
			valueExpr,
			valueExpr,
		)
	case "sql.NullBool":
		return fmt.Sprintf("sql.NullBool{Bool: %s, Valid: true}", valueExpr)
	case "sql.NullInt32":
		return fmt.Sprintf("sql.NullInt32{Int32: %s, Valid: true}", valueExpr)
	case "sql.NullInt64":
		return fmt.Sprintf("sql.NullInt64{Int64: %s, Valid: true}", valueExpr)
	case "sql.NullFloat64":
		return fmt.Sprintf(
			"sql.NullFloat64{Float64: %s, Valid: true}",
			valueExpr,
		)
	case "pgtype.Numeric":
		return fmt.Sprintf("func() pgtype.Numeric { var n pgtype.Numeric; _ = n.Scan(%s); return n }()", valueExpr)
	default:
		return valueExpr
	}
}

func normalizeSQLType(sqlType string) string {
	// Remove parameters like (255) or (10,2)
	normalizedType := strings.ToLower(sqlType)

	// Handle types with parameters
	if idx := strings.Index(normalizedType, "("); idx != -1 {
		normalizedType = normalizedType[:idx]
	}

	// Handle common aliases
	switch normalizedType {
	case "int4":
		return "integer"
	case "int8":
		return "bigint"
	case "int2":
		return "smallint"
	case "float4":
		return "real"
	case "float8":
		return "double precision"
	case "bool":
		return "boolean"
	}

	return normalizedType
}

func FormatFieldName(dbColumnName string) string {
	if dbColumnName == "id" {
		return "ID"
	}

	parts := strings.Split(dbColumnName, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}

	return strings.Join(parts, "")
}

func GenerateStructTags(column *catalog.Column, config TagConfig) string {
	var tags []string

	if config.JSON {
		tags = append(tags, fmt.Sprintf("json:\"%s\"", column.Name))
	}

	if config.DB {
		tags = append(tags, fmt.Sprintf("db:\"%s\"", column.Name))
	}

	if config.Validate {
		validationTags := generateValidationTags(column)
		if validationTags != "" {
			tags = append(tags, fmt.Sprintf("validate:\"%s\"", validationTags))
		}
	}

	for key, value := range config.Custom {
		tags = append(tags, fmt.Sprintf("%s:\"%s\"", key, value))
	}

	if len(tags) > 0 {
		return "`" + strings.Join(tags, " ") + "`"
	}

	return ""
}

func generateValidationTags(column *catalog.Column) string {
	var tags []string

	if !column.IsNullable {
		tags = append(tags, "required")
	}

	if strings.Contains(strings.ToLower(column.Name), "email") {
		tags = append(tags, "email")
	}

	if column.DataType == "uuid" {
		tags = append(tags, "uuid")
	}

	return strings.Join(tags, ",")
}

func (tm *TypeMapper) GenerateZeroCheck(field GeneratedField, valueExpr string) string {
	switch field.Type {
	case "string":
		return fmt.Sprintf("%s != \"\"", valueExpr)
	case "time.Time":
		return fmt.Sprintf("!%s.IsZero()", valueExpr)
	case "bool":
		return "true" // bools don't have meaningful zero checks in updates
	case "int32", "int64", "float32", "float64":
		return fmt.Sprintf("%s != 0", valueExpr)
	case "uuid.UUID":
		return fmt.Sprintf("%s != uuid.Nil", valueExpr)
	case "[]byte":
		return fmt.Sprintf("len(%s) > 0", valueExpr)
	default:
		return "true" // fallback
	}
}

type TagConfig struct {
	JSON     bool
	DB       bool
	Validate bool
	Custom   map[string]string
}
