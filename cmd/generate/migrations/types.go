package migrations

type Migration struct {
	FilePath   string
	Sequence   int
	Name       string
	Format     MigrationFormat
	UpSQL      string
	DownSQL    string
	Statements []string
}

type MigrationFormat int

const (
	GolangMigrate MigrationFormat = iota
	Goose
	Dbmate
)

func (f MigrationFormat) String() string {
	switch f {
	case GolangMigrate:
		return "golang-migrate"
	case Goose:
		return "goose"
	case Dbmate:
		return "dbmate"
	default:
		return "unknown"
	}
}