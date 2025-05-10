package seeds

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mbvlabs/grafto/models"
)

type tokenSeedData struct {
	ID         uuid.UUID
	CreatedAt  time.Time
	Expiration time.Time
	Meta       models.MetaInformation
	Type       string
}

type tokenSeedOption func(*tokenSeedData)

func WithTokenID(id uuid.UUID) tokenSeedOption {
	return func(tsd *tokenSeedData) {
		tsd.ID = id
	}
}

func WithTokenCreatedAt(createdAt time.Time) tokenSeedOption {
	return func(tsd *tokenSeedData) {
		tsd.CreatedAt = createdAt
	}
}

func WithTokenExpiration(expiration time.Time) tokenSeedOption {
	return func(tsd *tokenSeedData) {
		tsd.Expiration = expiration
	}
}

func WithTokenMeta(meta models.MetaInformation) tokenSeedOption {
	return func(tsd *tokenSeedData) {
		tsd.Meta = meta
	}
}

func WithHashedToken() tokenSeedOption {
	return func(tsd *tokenSeedData) {
		tsd.Type = "HASHED"
	}
}

func WithCodeToken() tokenSeedOption {
	return func(tsd *tokenSeedData) {
		tsd.Type = "CODE"
	}
}

func (s Seeder) PlantToken(
	ctx context.Context,
	opts ...tokenSeedOption,
) (models.Token, error) {
	data := &tokenSeedData{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		Expiration: time.Now().
			Add(24 * time.Hour),
		// Default expiration 24 hours from now
		Meta: models.MetaInformation{
			Resource:   models.ResourceUser,
			ResourceID: uuid.New(),
			Scope:      models.ScopeEmailVerification,
		},
		Type: "REGULAR",
	}

	for _, opt := range opts {
		opt(data)
	}

	var tkn models.Token

	switch data.Type {
	case "REGULAR":
		token, err := models.NewToken(ctx, s.dbtx, models.NewTokenPayload{
			Expiration: data.Expiration,
			Meta:       data.Meta,
		})
		if err != nil {
			return models.Token{}, err
		}
		tkn = token
	case "HASHED":
		token, err := models.NewHashedToken(ctx, s.dbtx, models.NewTokenPayload{
			Expiration: data.Expiration,
			Meta:       data.Meta,
		})
		if err != nil {
			return models.Token{}, err
		}
		tkn = token
	case "CODE":
		token, err := models.NewCodeToken(ctx, s.dbtx, models.NewTokenPayload{
			Expiration: data.Expiration,
			Meta:       data.Meta,
		})
		if err != nil {
			return models.Token{}, err
		}
		tkn = token
	}

	return tkn, nil
}

func (s Seeder) PlantTokens(
	ctx context.Context,
	amount int,
) ([]models.Token, error) {
	tokens := make([]models.Token, amount)

	for i := range amount {
		tkn, err := s.PlantToken(ctx)
		if err != nil {
			return nil, err
		}

		tokens[i] = tkn
	}

	return tokens, nil
}
