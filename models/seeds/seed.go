package seeds

import (
	"github.com/mbvlabs/grafto/models/internal/db"
)

type Seeder struct {
	dbtx db.DBTX
}

func NewSeeder(dbtx db.DBTX) Seeder {
	return Seeder{dbtx}
}
