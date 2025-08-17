package generator

import (
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/mbvlabs/grafto/cmd/generate/catalog"
)

type GeneratorConfig struct {
	TableName    string
	ResourceName string
	PackageName  string
	DatabaseType string
	StructTags   TagConfig
	CustomTypes  []TypeOverride
}

func GenerateModel(cat *catalog.Catalog, config GeneratorConfig) (*GeneratedModel, error) {
	table, err := cat.GetTable("", config.TableName)
	if err != nil {
		return nil, fmt.Errorf("table %s not found: %w", config.TableName, err)
	}
	
	typeMapper := NewTypeMapper(config.DatabaseType)
	for _, override := range config.CustomTypes {
		typeMapper.Overrides = append(typeMapper.Overrides, override)
	}
	
	model := &GeneratedModel{
		Name:      config.ResourceName,
		Package:   config.PackageName,
		TableName: config.TableName,
		Fields:    make([]GeneratedField, 0, len(table.Columns)),
		Imports:   make([]string, 0),
	}
	
	// Track required imports
	importSet := make(map[string]bool)
	
	for _, col := range table.Columns {
		goType, sqlcType, pkg, err := typeMapper.MapSQLTypeToGo(col.DataType, col.IsNullable)
		if err != nil {
			return nil, fmt.Errorf("failed to map type for column %s: %w", col.Name, err)
		}
		
		field := GeneratedField{
			Name:     FormatFieldName(col.Name),
			Type:     goType,
			SQLCType: sqlcType,
			Tag:      GenerateStructTags(col, config.StructTags),
		}
		
		field.ConversionFromDB = typeMapper.GenerateConversionFromDB(field)
		// For insert operations, use different value expressions based on field type
		if col.Name == "created_at" || col.Name == "updated_at" {
			field.ConversionToDB = typeMapper.GenerateConversionToDB(field, "resource."+field.Name)
		} else {
			field.ConversionToDB = typeMapper.GenerateConversionToDB(field, "data."+field.Name)
		}
		
		if pkg != "" {
			importSet[pkg] = true
		}
		
		// Add imports for nullable types
		if col.IsNullable {
			switch sqlcType {
			case "sql.NullString", "sql.NullBool", "sql.NullInt32", "sql.NullInt64", "sql.NullFloat64":
				importSet["database/sql"] = true
			case "pgtype.Timestamptz", "pgtype.Numeric":
				importSet["github.com/jackc/pgx/v5/pgtype"] = true
			}
		}
		
		model.Fields = append(model.Fields, field)
	}
	
	// Convert import set to sorted slice
	for imp := range importSet {
		model.Imports = append(model.Imports, imp)
	}
	sort.Strings(model.Imports)
	
	return model, nil
}

func GenerateModelFile(model *GeneratedModel, templateStr string) (string, error) {
	tmpl, err := template.New("model").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}
	
	var buf strings.Builder
	if err := tmpl.Execute(&buf, model); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	
	return buf.String(), nil
}

func GetDefaultModelTemplate() string {
	return `package {{.Package}}

import (
{{- range .Imports}}
	"{{.}}"
{{- end}}
	"context"
	"errors"

	"github.com/mbvlabs/grafto/models/internal/db"
)

type {{.Name}} struct {
{{- range .Fields}}
	{{.Name}} {{.Type}} {{.Tag}}
{{- end}}
}

type New{{.Name}}Payload struct {
{{- range .Fields}}
{{- if and (ne .Name "ID") (ne .Name "CreatedAt") (ne .Name "UpdatedAt")}}
	{{.Name}} {{.Type}} {{.Tag}}
{{- end}}
{{- end}}
}

func New{{.Name}}(
	ctx context.Context,
	dbtx db.DBTX,
	data New{{.Name}}Payload,
) ({{.Name}}, error) {
	if err := validate.Struct(data); err != nil {
		return {{.Name}}{}, errors.Join(ErrDomainValidation, err)
	}

	resource := {{.Name}}{
{{- range .Fields}}
{{- if eq .Name "ID"}}
		{{.Name}}: uuid.New(),
{{- else if or (eq .Name "CreatedAt") (eq .Name "UpdatedAt")}}
		{{.Name}}: time.Now(),
{{- else}}
		{{.Name}}: data.{{.Name}},
{{- end}}
{{- end}}
	}

	row, err := db.Stmts.Insert{{.Name}}(ctx, dbtx, db.Insert{{.Name}}Params{
{{- range .Fields}}
{{- if eq .Name "ID"}}
		{{.Name}}: resource.{{.Name}},
{{- else if or (eq .Name "CreatedAt") (eq .Name "UpdatedAt")}}
		{{.Name}}: {{.ConversionToDB}},
{{- else}}
		{{.Name}}: {{.ConversionToDB}},
{{- end}}
{{- end}}
	})
	if err != nil {
		return {{.Name}}{}, err
	}

	return rowTo{{.Name}}(row), nil
}

func Find{{.Name}}(
	ctx context.Context,
	dbtx db.DBTX,
	id uuid.UUID,
) ({{.Name}}, error) {
	row, err := db.Stmts.Query{{.Name}}ByID(ctx, dbtx, id)
	if err != nil {
		return {{.Name}}{}, err
	}

	return rowTo{{.Name}}(row), nil
}

type Update{{.Name}}Payload struct {
	ID uuid.UUID ` + "`validate:\"required,uuid\"`" + `
{{- range .Fields}}
{{- if and (ne .Name "ID") (ne .Name "CreatedAt")}}
	{{.Name}} {{.Type}}
{{- end}}
{{- end}}
}

func Update{{.Name}}(
	ctx context.Context,
	dbtx db.DBTX,
	data Update{{.Name}}Payload,
) ({{.Name}}, error) {
	if err := validate.Struct(data); err != nil {
		return {{.Name}}{}, errors.Join(ErrDomainValidation, err)
	}

	currentRow, err := db.Stmts.Query{{.Name}}ByID(ctx, dbtx, data.ID)
	if err != nil {
		return {{.Name}}{}, err
	}

	payload := db.Update{{.Name}}Params{
		ID: data.ID,
		UpdatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
{{- range .Fields}}
{{- if and (ne .Name "ID") (ne .Name "CreatedAt") (ne .Name "UpdatedAt")}}
		{{.Name}}: currentRow.{{.Name}},
{{- end}}
{{- end}}
	}

{{- range .Fields}}
{{- if and (ne .Name "ID") (ne .Name "CreatedAt") (ne .Name "UpdatedAt")}}
	if !isEmpty(data.{{.Name}}) {
		payload.{{.Name}} = {{.ConversionToDB}}
	}
{{- end}}
{{- end}}

	row, err := db.Stmts.Update{{.Name}}(ctx, dbtx, payload)
	if err != nil {
		return {{.Name}}{}, err
	}

	return rowTo{{.Name}}(row), nil
}

func Delete{{.Name}}(
	ctx context.Context,
	dbtx db.DBTX,
	id uuid.UUID,
) error {
	return db.Stmts.Delete{{.Name}}(ctx, dbtx, id)
}

func rowTo{{.Name}}(row db.{{.Name}}) {{.Name}} {
	return {{.Name}}{
{{- range .Fields}}
		{{.Name}}: {{.ConversionFromDB}},
{{- end}}
	}
}

func isEmpty(v interface{}) bool {
	switch val := v.(type) {
	case string:
		return val == ""
	case time.Time:
		return val.IsZero()
	case bool:
		return false
	case int32, int64, float64, float32:
		return val == 0
	case uuid.UUID:
		return val == uuid.Nil
	default:
		return false
	}
}

// Helper functions for pgtype conversions
func numericToFloat64(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0
	}
	f, _ := n.Float64Value()
	return f.Float64
}

func float64ToNumeric(f float64) pgtype.Numeric {
	var n pgtype.Numeric
	_ = n.Scan(f)
	return n
}

type Paginated{{.Name}}s struct {
	{{.Name}}s    []{{.Name}}
	TotalCount int64
	Page       int64
	PageSize   int64
	TotalPages int64
}

func GetPaginated{{.Name}}s(
	ctx context.Context,
	dbtx db.DBTX,
	page int64,
	pageSize int64,
) (Paginated{{.Name}}s, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	totalCount, err := db.Stmts.Count{{.Name}}s(ctx, dbtx)
	if err != nil {
		return Paginated{{.Name}}s{}, err
	}

	rows, err := db.Stmts.QueryPaginated{{.Name}}s(
		ctx,
		dbtx,
		db.QueryPaginated{{.Name}}sParams{
			Limit:  pageSize,
			Offset: offset,
		},
	)
	if err != nil {
		return Paginated{{.Name}}s{}, err
	}

	{{.TableName}} := make([]{{.Name}}, len(rows))
	for i, row := range rows {
		{{.TableName}}[i] = rowTo{{.Name}}(row)
	}

	totalPages := (totalCount + int64(pageSize) - 1) / int64(pageSize)

	return Paginated{{.Name}}s{
		{{.Name}}s:    {{.TableName}},
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}`
}