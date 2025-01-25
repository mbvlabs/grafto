package seeds

import (
	"github.com/mbvlabs/grafto/models/internal/db"
	"golang.org/x/net/context"
)

type SeedBuilder[T, V any] interface {
	WithRandoms(n int) *T
	WithSpecific(data map[string]any) *T
	Build() *V
}

type Seed interface {
	Generate(ctx context.Context, dbtx db.DBTX) error
}

type Seeder struct {
	dbtx db.DBTX
}

func NewSeeder() Seeder {
	return Seeder{}
}
