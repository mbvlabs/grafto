package seeds

import (
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mbvlabs/grafto/models"
	"github.com/mbvlabs/grafto/models/internal/db"
	"golang.org/x/net/context"
)

type userSeedData struct {
	ID              uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Email           string
	EmailVerifiedAt time.Time
	IsAdmin         bool
}

type userSeedOption func(*userSeedData)

func WithUserID(id uuid.UUID) userSeedOption {
	return func(usd *userSeedData) {
		usd.ID = id
	}
}

func WithUserCreatedAt(createdAt time.Time) userSeedOption {
	return func(usd *userSeedData) {
		usd.CreatedAt = createdAt
	}
}

func WithUserUpdatedAt(updatedAt time.Time) userSeedOption {
	return func(usd *userSeedData) {
		usd.UpdatedAt = updatedAt
	}
}

func WithUserEmail(email string) userSeedOption {
	return func(usd *userSeedData) {
		usd.Email = email
	}
}

func WithUserEmailVerifiedAt(emailVerifiedAt time.Time) userSeedOption {
	return func(usd *userSeedData) {
		usd.EmailVerifiedAt = emailVerifiedAt
	}
}

func WithUserIsAdmin(isAdmin bool) userSeedOption {
	return func(usd *userSeedData) {
		usd.IsAdmin = isAdmin
	}
}

func (s Seeder) PlantUser(
	ctx context.Context,
	opts ...userSeedOption,
) (models.UserEntity, error) {
	data := &userSeedData{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Email:     faker.Email(),
	}

	for _, opt := range opts {
		opt(data)
	}

	user, err := models.NewUser(ctx, s.dbtx, models.NewUserPayload{
		Email: data.Email,
		Password: models.PasswordPair{
			Password:        "password",
			ConfirmPassword: "password",
		},
	})
	if err != nil {
		return models.UserEntity{}, err
	}

	if !data.EmailVerifiedAt.IsZero() {
		if err := db.Stmts.VerifyUserEmail(ctx, s.dbtx, db.VerifyUserEmailParams{
			Email: data.Email,
			UpdatedAt: pgtype.Timestamptz{
				Time:  data.UpdatedAt,
				Valid: true,
			},
			EmailVerifiedAt: pgtype.Timestamptz{
				Time:  data.EmailVerifiedAt,
				Valid: true,
			},
		}); err != nil {
			return models.UserEntity{}, err
		}
	}

	if data.IsAdmin {
		if _, err := db.Stmts.UpdateUserIsAdmin(ctx, s.dbtx, db.UpdateUserIsAdminParams{
			ID:      user.ID,
			IsAdmin: data.IsAdmin,
			UpdatedAt: pgtype.Timestamptz{
				Time:  data.UpdatedAt,
				Valid: true,
			},
		}); err != nil {
			return models.UserEntity{}, err
		}
	}

	return user, nil
}

func (s Seeder) PlantUsers(
	ctx context.Context,
	amount int,
) ([]models.UserEntity, error) {
	users := make([]models.UserEntity, amount)

	for i := range amount {
		usr, err := s.PlantUser(ctx)
		if err != nil {
			return nil, err
		}

		users[i] = usr
	}

	return users, nil
}
