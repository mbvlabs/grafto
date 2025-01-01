package seeds

import (
	"context"
	"math/rand"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mbvlabs/grafto/models"
	"github.com/mbvlabs/grafto/models/internal/db"
	"github.com/mbvlabs/grafto/services"
)

type UserSeed struct {
	entities []models.UserEntity
}

type UserSeedBuilder struct {
	seed *UserSeed
}

var (
	_ Seed                                   = (*UserSeed)(nil)
	_ SeedBuilder[UserSeedBuilder, UserSeed] = (*UserSeedBuilder)(nil)
)

func NewUserSeedBuilder() *UserSeedBuilder {
	return &UserSeedBuilder{
		seed: &UserSeed{
			entities: make([]models.UserEntity, 0),
		},
	}
}

func (b *UserSeedBuilder) WithRandoms(n int) *UserSeedBuilder {
	for i := 0; i < n; i++ {
		trueOrFalse := rand.Float32() < 0.5
		var emailVerifiedAt time.Time
		if trueOrFalse {
			emailVerifiedAt = time.Now()
		}

		b.seed.entities = append(b.seed.entities, models.UserEntity{
			ID:              uuid.New(),
			CreatedAt:       time.Now().Add(time.Minute * time.Duration(i)),
			UpdatedAt:       time.Now().Add(time.Minute * time.Duration(i)),
			Email:           faker.Email(),
			EmailVerifiedAt: emailVerifiedAt,
			IsAdmin:         trueOrFalse,
		})
	}
	return b
}

// WithUser adds a specific user to the seed
func (b *UserSeedBuilder) WithSpecific(data map[string]any) *UserSeedBuilder {
	entity := models.UserEntity{
		ID:              uuid.New(),
		CreatedAt:       time.Now().Add(time.Minute * time.Duration(1)),
		UpdatedAt:       time.Now().Add(time.Minute * time.Duration(2)),
		Email:           faker.Email(),
		EmailVerifiedAt: time.Now(),
		IsAdmin:         false,
	}
	for k, v := range data {
		switch k {
		case "ID":
			entity.ID = v.(uuid.UUID)
		case "CreatedAt":
			entity.CreatedAt = v.(time.Time)
		case "UpdatedAt":
			entity.UpdatedAt = v.(time.Time)
		case "Email":
			entity.Email = v.(string)
		case "EmailVerifiedAt":
			entity.EmailVerifiedAt = v.(time.Time)
		case "IsAdmin":
			entity.IsAdmin = v.(bool)
		}
	}

	b.seed.entities = append(b.seed.entities, entity)
	return b
}

// Build returns the final UserSeed
func (b *UserSeedBuilder) Build() *UserSeed {
	return b.seed
}

// Generate implements the Seeder interface
func (s *UserSeed) Generate(ctx context.Context, dbtx db.DBTX) error {
	hashedBytes, err := services.HashAndPepperPassword("password")
	if err != nil {
		return err
	}

	for _, user := range s.entities {
		_, err := db.Stmts.InsertUser(ctx, dbtx, db.InsertUserParams{
			ID:        user.ID,
			CreatedAt: pgtype.Timestamptz{Time: user.CreatedAt, Valid: true},
			UpdatedAt: pgtype.Timestamptz{Time: user.UpdatedAt, Valid: true},
			Email:     user.Email,
			Password:  string(hashedBytes),
			IsAdmin:   user.IsAdmin,
		})
		if err != nil {
			return err
		}

		if !user.EmailVerifiedAt.IsZero() {
			if err := db.Stmts.VerifyUserEmail(ctx, dbtx, db.VerifyUserEmailParams{
				Email: user.Email,
				UpdatedAt: pgtype.Timestamptz{
					Time:  user.UpdatedAt,
					Valid: true,
				},
				EmailVerifiedAt: pgtype.Timestamptz{
					Time:  user.EmailVerifiedAt,
					Valid: true,
				},
			}); err != nil {
				return err
			}
		}
	}

	return nil
}
