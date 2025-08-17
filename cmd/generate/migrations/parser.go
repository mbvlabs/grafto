package migrations

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// DiscoverMigrations finds all .sql files in given directories
func DiscoverMigrations(dirs []string) ([]Migration, error) {
	var migrations []Migration

	for _, dir := range dirs {
		dirMigrations, err := discoverMigrationsInDir(dir)
		if err != nil {
			return nil, fmt.Errorf("failed to discover migrations in %s: %w", dir, err)
		}
		migrations = append(migrations, dirMigrations...)
	}

	// Sort migrations lexicographically by filename
	sort.Slice(migrations, func(i, j int) bool {
		return filepath.Base(migrations[i].FilePath) < filepath.Base(migrations[j].FilePath)
	})

	return migrations, nil
}

func discoverMigrationsInDir(dir string) ([]Migration, error) {
	var migrations []Migration

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".sql") {
			return nil
		}

		if IsDownMigration(filepath.Base(path)) {
			return nil
		}

		migration, err := ParseMigration(path)
		if err != nil {
			return fmt.Errorf("failed to parse migration %s: %w", path, err)
		}

		migrations = append(migrations, *migration)
		return nil
	})

	return migrations, err
}

// ParseMigration extracts DDL statements from migration file
func ParseMigration(filePath string) (*Migration, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read migration file: %w", err)
	}

	filename := filepath.Base(filePath)
	sequence, name, err := parseFilename(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to parse filename %s: %w", filename, err)
	}

	format := detectMigrationFormat(string(content))
	upSQL := RemoveRollbackStatements(string(content), format)
	downSQL := extractDownSQL(string(content), format)

	migration := &Migration{
		FilePath:   filePath,
		Sequence:   sequence,
		Name:       name,
		Format:     format,
		UpSQL:      upSQL,
		DownSQL:    downSQL,
		Statements: parseStatements(upSQL),
	}

	return migration, nil
}

// RemoveRollbackStatements filters out "down" migration content
func RemoveRollbackStatements(content string, format MigrationFormat) string {
	switch format {
	case GolangMigrate:
		return extractUpSQLGolangMigrate(content)
	case Goose:
		return extractUpSQLGoose(content)
	case Dbmate:
		return extractUpSQLDbmate(content)
	default:
		return content
	}
}

// IsDownMigration checks if file is a rollback migration
func IsDownMigration(filename string) bool {
	return strings.Contains(filename, ".down.") || strings.HasSuffix(filename, ".down.sql")
}

func parseFilename(filename string) (sequence int, name string, err error) {
	// Extract sequence number from filename
	// Supports formats like: 00001_name.sql, 001_name.sql, 1_name.sql
	re := regexp.MustCompile(`^(\d+)_(.+)\.sql$`)
	matches := re.FindStringSubmatch(filename)

	if len(matches) != 3 {
		return 0, "", fmt.Errorf("invalid migration filename format: %s", filename)
	}

	sequence, err = strconv.Atoi(matches[1])
	if err != nil {
		return 0, "", fmt.Errorf("invalid sequence number in filename: %s", matches[1])
	}

	name = matches[2]
	return sequence, name, nil
}

func detectMigrationFormat(content string) MigrationFormat {
	if strings.Contains(content, "-- migrate:up") || strings.Contains(content, "-- migrate:down") {
		return GolangMigrate
	}
	if strings.Contains(content, "-- +goose Up") || strings.Contains(content, "-- +goose Down") {
		return Goose
	}
	if strings.Contains(content, "-- migrate:up") || strings.Contains(content, "-- migrate:down") {
		return Dbmate
	}
	// Default to golang-migrate if no specific markers found
	return GolangMigrate
}

func extractUpSQLGolangMigrate(content string) string {
	lines := strings.Split(content, "\n")
	var upLines []string
	inUp := false

	for _, line := range lines {
		if strings.TrimSpace(line) == "-- migrate:up" {
			inUp = true
			continue
		}
		if strings.TrimSpace(line) == "-- migrate:down" {
			break
		}
		if inUp {
			upLines = append(upLines, line)
		}
	}

	if !inUp {
		// If no explicit up section, assume entire content is up
		return content
	}

	return strings.Join(upLines, "\n")
}

func extractUpSQLGoose(content string) string {
	lines := strings.Split(content, "\n")
	var upLines []string
	inUp := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "-- +goose Up") {
			inUp = true
			continue
		}
		if strings.HasPrefix(trimmed, "-- +goose Down") {
			break
		}
		if inUp && !strings.HasPrefix(trimmed, "-- +goose StatementBegin") && !strings.HasPrefix(trimmed, "-- +goose StatementEnd") {
			if !strings.HasPrefix(trimmed, "SELECT ") || !strings.Contains(trimmed, "SQL query") {
				upLines = append(upLines, line)
			}
		}
	}

	return strings.Join(upLines, "\n")
}

func extractUpSQLDbmate(content string) string {
	// Dbmate uses same format as golang-migrate
	return extractUpSQLGolangMigrate(content)
}

func extractDownSQL(content string, format MigrationFormat) string {
	switch format {
	case GolangMigrate:
		return extractDownSQLGolangMigrate(content)
	case Goose:
		return extractDownSQLGoose(content)
	case Dbmate:
		return extractDownSQLDbmate(content)
	default:
		return ""
	}
}

func extractDownSQLGolangMigrate(content string) string {
	lines := strings.Split(content, "\n")
	var downLines []string
	inDown := false

	for _, line := range lines {
		if strings.TrimSpace(line) == "-- migrate:down" {
			inDown = true
			continue
		}
		if inDown {
			downLines = append(downLines, line)
		}
	}

	return strings.Join(downLines, "\n")
}

func extractDownSQLGoose(content string) string {
	lines := strings.Split(content, "\n")
	var downLines []string
	inDown := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "-- +goose Down") {
			inDown = true
			continue
		}
		if inDown && !strings.HasPrefix(trimmed, "-- +goose StatementBegin") && !strings.HasPrefix(trimmed, "-- +goose StatementEnd") {
			if !strings.HasPrefix(trimmed, "SELECT ") || !strings.Contains(trimmed, "SQL query") {
				downLines = append(downLines, line)
			}
		}
	}

	return strings.Join(downLines, "\n")
}

func extractDownSQLDbmate(content string) string {
	return extractDownSQLGolangMigrate(content)
}

func parseStatements(sql string) []string {
	var statements []string

	// Split by semicolon but handle string literals and comments
	lines := strings.Split(sql, "\n")
	var currentStatement strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}

		currentStatement.WriteString(line)
		currentStatement.WriteString("\n")

		// Check if statement ends with semicolon
		if strings.HasSuffix(trimmed, ";") {
			stmt := strings.TrimSpace(currentStatement.String())
			if stmt != "" {
				statements = append(statements, stmt)
			}
			currentStatement.Reset()
		}
	}

	// Handle case where last statement doesn't end with semicolon
	if currentStatement.Len() > 0 {
		stmt := strings.TrimSpace(currentStatement.String())
		if stmt != "" {
			statements = append(statements, stmt)
		}
	}

	return statements
}
