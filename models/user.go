package models

import (
	"context"
	"crypto/subtle"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mbvlabs/grafto/config"
	"github.com/mbvlabs/grafto/models/internal/db"
	"golang.org/x/crypto/argon2"
)

type UserEntity struct {
	ID              uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Email           string
	EmailVerifiedAt time.Time
	IsAdmin         bool
	HashedPassword  string
}

func (ue UserEntity) IsVerified() bool {
	return !ue.EmailVerifiedAt.IsZero()
}

func (ue UserEntity) ValidatePassword(providedPassword string) error {
	if t := subtle.ConstantTimeCompare(HashPassword(providedPassword), []byte(ue.HashedPassword)); t == 1 {
		return nil
	}

	return errors.New("invalid password")
}

type PasswordPair struct {
	Password        string `validate:"required,gte=6"`
	ConfirmPassword string `validate:"required,gte=6"`
}

func HashPassword(password string) []byte {
	return argon2.IDKey(
		[]byte(password),
		[]byte(config.Cfg.PasswordSalt),
		2,
		19*1024,
		1,
		32,
	)
}

func GetUserByEmail(
	ctx context.Context,
	dbtx db.DBTX,
	email string,
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
		HashedPassword:  string(user.Password),
		EmailVerifiedAt: user.EmailVerifiedAt.Time,
		IsAdmin:         user.IsAdmin,
	}, nil
}

func GetUser(
	ctx context.Context,
	dbtx db.DBTX,
	id uuid.UUID,
) (UserEntity, error) {
	row, err := db.Stmts.QueryUserByID(ctx, dbtx, id)
	if err != nil {
		return UserEntity{}, err
	}

	return UserEntity{
		ID:              id,
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
		Email:           row.Email,
		EmailVerifiedAt: row.EmailVerifiedAt.Time,
		HashedPassword:  string(row.Password),
		IsAdmin:         false,
	}, nil
}

type NewUserPayload struct {
	Email    string `validate:"required,email"`
	Password PasswordPair
}

func NewUser(
	ctx context.Context,
	dbtx db.DBTX,
	data NewUserPayload,
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
	hp := HashPassword(data.Password.Password)
	usr.HashedPassword = string(hp)

	_, err := db.Stmts.InsertUser(ctx, dbtx, db.InsertUserParams{
		ID:        usr.ID,
		CreatedAt: pgtype.Timestamptz{Time: usr.CreatedAt, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: usr.UpdatedAt, Valid: true},
		Email:     usr.Email,
		Password:  hp,
	})
	if err != nil {
		return UserEntity{}, err
	}

	return usr, nil
}

type UpdateUserPayload struct {
	ID        uuid.UUID `validate:"required,uuid"`
	UpdatedAt time.Time `validate:"required"`
	Email     string    `validate:"required,email"`
}

func UpdateUser(
	ctx context.Context,
	dbtx db.DBTX,
	data UpdateUserPayload,
) (UserEntity, error) {
	if err := validate.Struct(data); err != nil {
		return UserEntity{}, errors.Join(ErrDomainValidation, err)
	}

	row, err := db.Stmts.UpdateUser(ctx, dbtx, db.UpdateUserParams{
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
		ID:              row.ID,
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
		Email:           row.Email,
		EmailVerifiedAt: row.EmailVerifiedAt.Time,
		IsAdmin:         false,
	}, nil
}

type UpdateUserPasswordPayload struct {
	ID        uuid.UUID `validate:"required,uuid"`
	UpdatedAt time.Time `validate:"required"`
	Password  PasswordPair
}

func UpdateUserPassword(
	ctx context.Context,
	dbtx db.DBTX,
	data UpdateUserPasswordPayload,
) error {
	if err := validate.Struct(data); err != nil {
		return errors.Join(ErrDomainValidation, err)
	}

	return db.Stmts.ChangeUserPassword(ctx, dbtx, db.ChangeUserPasswordParams{
		ID: data.ID,
		UpdatedAt: pgtype.Timestamptz{
			Time:  data.UpdatedAt,
			Valid: true,
		},
		Password: HashPassword(data.Password.Password),
	})
}

type UpdateUserEmailToVerifiedPayload struct {
	ID         uuid.UUID `validate:"required,uuid"`
	Email      string    `validate:"required,email"`
	VerifiedAt time.Time `validate:"required"`
}

func UpdateUserEmailToVerified(
	ctx context.Context,
	dbtx db.DBTX,
	data UpdateUserEmailToVerifiedPayload,
) error {
	if err := validate.Struct(data); err != nil {
		return errors.Join(ErrDomainValidation, err)
	}

	time := pgtype.Timestamptz{
		Time:  data.VerifiedAt,
		Valid: true,
	}

	return db.Stmts.VerifyUserEmail(ctx, dbtx, db.VerifyUserEmailParams{
		Email:           data.Email,
		UpdatedAt:       time,
		EmailVerifiedAt: time,
	})
}

type MakeUserAdminPayload struct {
	UserID    uuid.UUID `validate:"required,uuid"`
	UpdatedAt time.Time `validate:"required"`
	ActorID   uuid.UUID `validate:"required,uuid"`
}

func MakeUserAdmin(
	ctx context.Context,
	dbtx db.DBTX,
	data MakeUserAdminPayload,
) (UserEntity, error) {
	if err := validate.Struct(data); err != nil {
		return UserEntity{}, errors.Join(ErrDomainValidation, err)
	}

	actor, err := GetUser(ctx, dbtx, data.ActorID)
	if err != nil {
		return UserEntity{}, err
	}
	if !actor.IsAdmin {
		return UserEntity{}, ErrMustBeAdmin
	}

	row, err := db.Stmts.UpdateUserIsAdmin(
		ctx,
		dbtx,
		db.UpdateUserIsAdminParams{
			ID:        data.UserID,
			IsAdmin:   true,
			UpdatedAt: pgtype.Timestamptz{Time: data.UpdatedAt, Valid: true},
		},
	)
	if err != nil {
		return UserEntity{}, err
	}

	return UserEntity{
		ID:              row.ID,
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
		Email:           row.Email,
		EmailVerifiedAt: row.EmailVerifiedAt.Time,
		IsAdmin:         true,
	}, nil
}
