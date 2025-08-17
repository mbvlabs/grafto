package config

import "github.com/mbvlabs/grafto/cmd/generate/generator"

type Config struct {
	MigrationDirs []string
	DatabaseType  string
	TableName     string
	OutputFile    string
	PackageName   string
	GenerateJSON  bool
	CustomTypes   []generator.TypeOverride
	StructTags    generator.TagConfig
}

func NewDefaultConfig() *Config {
	return &Config{
		MigrationDirs: []string{"psql/migrations"},
		DatabaseType:  "postgresql",
		PackageName:   "models",
		GenerateJSON:  true,
		StructTags: generator.TagConfig{
			JSON:     true,
			DB:       true,
			Validate: true,
			Custom:   make(map[string]string),
		},
	}
}