package models

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mbvlabs/grafto/models/internal/db"
)

type UserEntity struct {
	ID              uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Email           string
	EmailVerifiedAt time.Time
	IsAdmin         bool
}

func (ue UserEntity) IsVerified() bool {
	return !ue.EmailVerifiedAt.IsZero()
}

func GetUserByEmail(
	ctx context.Context,
	email string,
	dbtx db.DBTX,
) (UserEntity, error) {
	user, err := db.Stmts.QueryUserByEmail(ctx, dbtx, email)
	if err != nil {
		return UserEntity{}, err
	}

	return UserEntity{
		ID:              user.ID,
		CreatedAt:       user.CreatedAt.Time,
		UpdatedAt:       user.UpdatedAt.Time,
		Email:           user.Email,
		EmailVerifiedAt: user.EmailVerifiedAt.Time,
		IsAdmin:         user.IsAdmin,
	}, nil
}

type NewUserPayload struct {
	Email           string `validate:"required,email"`
	Password        string `validate:"required,gte=6"`
	ConfirmPassword string `validate:"required,gte=6"`
}

func GetUser(ctx context.Context, id uuid.UUID, dbtx db.DBTX) (UserEntity, error) {
	usr, err := db.Stmts.QueryUserByID(ctx, dbtx, id)
	if err != nil {
		return UserEntity{}, err
	}

	return UserEntity{
		ID:              id,
		CreatedAt:       usr.CreatedAt.Time,
		UpdatedAt:       usr.UpdatedAt.Time,
		Email:           usr.Email,
		EmailVerifiedAt: usr.EmailVerifiedAt.Time,
		IsAdmin:         false,
	}, nil
}

func NewUser(
	ctx context.Context,
	data NewUserPayload,
	dbtx db.DBTX,
	hash func(password string) (string, error),
) (UserEntity, error) {
	if err := validate.Struct(data); err != nil {
		return UserEntity{}, errors.Join(ErrDomainValidation, err)
	}

	usr := UserEntity{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Email:     data.Email,
	}

	hashedPassword, err := hash(data.Password)
	if err != nil {
		return UserEntity{}, err
	}

	_, err = db.Stmts.InsertUser(ctx, dbtx, db.InsertUserParams{
		ID:        usr.ID,
		CreatedAt: pgtype.Timestamptz{Time: usr.CreatedAt, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: usr.UpdatedAt, Valid: true},
		Email:     usr.Email,
		Password:  hashedPassword,
	})
	if err != nil {
		return UserEntity{}, err
	}

	return usr, nil
}

type UpdateUserPayload struct {
	ID             uuid.UUID `validate:"required,uuid"`
	UpdatedAt      time.Time `validate:"required"`
	Name           string    `validate:"required,gte=2,lte=25"`
	Email          string    `validate:"required,email"`
	EmailUpdatedAt time.Time
}

func UpdateUser(
	ctx context.Context,
	data UpdateUserPayload,
	dbtx db.DBTX,
) (UserEntity, error) {
	// validate payload
	if err := validate.Struct(data); err != nil {
		return UserEntity{}, errors.Join(ErrDomainValidation, err)
	}

	updatedUsr, err := db.Stmts.UpdateUser(ctx, dbtx, db.UpdateUserParams{
		ID: data.ID,
		UpdatedAt: pgtype.Timestamptz{
			Time:  data.UpdatedAt,
			Valid: true,
		},
		Email: data.Email,
	})
	if err != nil {
		return UserEntity{}, err
	}

	return UserEntity{
		ID:              updatedUsr.ID,
		CreatedAt:       updatedUsr.CreatedAt.Time,
		UpdatedAt:       updatedUsr.UpdatedAt.Time,
		Email:           updatedUsr.Email,
		EmailVerifiedAt: updatedUsr.EmailVerifiedAt.Time,
		IsAdmin:         false,
	}, nil
}

type UpdateUserPasswordPayload struct {
	ID        uuid.UUID `validate:"required,uuid"`
	UpdatedAt time.Time `validate:"required"`
	Password  string    `validate:"required"`
}

func UpdateUserPassword(
	ctx context.Context,
	data UpdateUserPasswordPayload,
	q func(
		ctx context.Context,
		userID uuid.UUID,
		newPassword string,
		updatedAt time.Time,
	) error,
) error {
	// validate payload
	if err := validate.Struct(data); err != nil {
		return errors.Join(ErrDomainValidation, err)
	}

	return q(ctx, data.ID, data.Password, data.UpdatedAt)
}
