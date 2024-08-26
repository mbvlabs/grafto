package psql

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mbvlabs/grafto/models"
	"github.com/mbvlabs/grafto/psql/database"
)

func (p Postgres) QueryUserByID(
	ctx context.Context,
	id uuid.UUID,
) (models.User, error) {
	user, err := p.Queries.QueryUserByID(ctx, id)
	if err != nil {
		return models.User{}, err
	}

	return models.User{
		ID:              user.ID,
		CreatedAt:       user.CreatedAt.Time,
		UpdatedAt:       user.UpdatedAt.Time,
		Name:            user.Name,
		Email:           user.Mail,
		EmailVerifiedAt: user.MailVerifiedAt.Time,
	}, nil
}

func (p Postgres) QueryUserByEmail(
	ctx context.Context,
	email string,
) (models.User, error) {
	user, err := p.Queries.QueryUserByEmail(ctx, email)
	if err != nil {
		return models.User{}, err
	}

	return models.User{
		ID:              user.ID,
		CreatedAt:       user.CreatedAt.Time,
		UpdatedAt:       user.UpdatedAt.Time,
		Name:            user.Name,
		Email:           user.Mail,
		EmailVerifiedAt: user.MailVerifiedAt.Time,
	}, nil
}

func (p Postgres) InsertUser(
	ctx context.Context,
	data models.User,
	hashedPassword string,
) (models.User, error) {
	createdAt := pgtype.Timestamptz{
		Time:  data.CreatedAt,
		Valid: true,
	}
	updatedAt := pgtype.Timestamptz{
		Time:  data.UpdatedAt,
		Valid: true,
	}

	_, err := p.Queries.InsertUser(ctx, database.InsertUserParams{
		ID:        data.ID,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Name:      data.Name,
		Mail:      data.Email,
		Password:  hashedPassword,
	})
	if err != nil {
		return models.User{}, err
	}

	return data, nil
}

func (p Postgres) UpdateUser(
	ctx context.Context,
	data models.User,
) (models.User, error) {
	updatedAt := pgtype.Timestamptz{
		Time:  data.UpdatedAt,
		Valid: true,
	}

	_, err := p.Queries.UpdateUser(ctx, database.UpdateUserParams{
		ID:        data.ID,
		UpdatedAt: updatedAt,
		Name:      data.Name,
		Mail:      data.Email,
	})
	if err != nil {
		return models.User{}, err
	}

	return data, nil
}

func (p Postgres) UpdateUserPassword(
	ctx context.Context,
	userID uuid.UUID,
	newPassword string,
	updatedAt time.Time,
) error {
	parsedUpdatedAt := pgtype.Timestamptz{
		Time:  updatedAt,
		Valid: true,
	}

	if err := p.Queries.ChangeUserPassword(ctx, database.ChangeUserPasswordParams{
		ID:        userID,
		Password:  newPassword,
		UpdatedAt: parsedUpdatedAt,
	}); err != nil {
		return err
	}

	return nil
}

func (p Postgres) VerifyUserEmail(
	ctx context.Context,
	updatedAt time.Time,
	email string,
) error {
	parsedUpdatedAt := pgtype.Timestamptz{
		Time:  updatedAt,
		Valid: true,
	}

	return p.Queries.VerifyUserEmail(ctx, database.VerifyUserEmailParams{
		Mail:           email,
		UpdatedAt:      parsedUpdatedAt,
		MailVerifiedAt: parsedUpdatedAt,
	})
}
