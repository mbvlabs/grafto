package models

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mbvlabs/grafto/config"
	"github.com/mbvlabs/grafto/models/internal/db"
	"golang.org/x/crypto/bcrypt"
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
	return bcrypt.CompareHashAndPassword(
		[]byte(ue.HashedPassword),
		[]byte(providedPassword+config.Cfg.PasswordPepper),
	)
}

func HashAndPepperPassword(password string) (string, error) {
	passwordBytes := []byte(password + config.Cfg.PasswordPepper)
	hashedBytes, err := bcrypt.GenerateFromPassword(
		passwordBytes,
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", err
	}

	return string(hashedBytes), nil
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
		HashedPassword:  user.Password,
		EmailVerifiedAt: user.EmailVerifiedAt.Time,
		IsAdmin:         user.IsAdmin,
	}, nil
}

type NewUserPayload struct {
	Email           string `validate:"required,email"`
	Password        string `validate:"required,gte=6"`
	ConfirmPassword string `validate:"required,gte=6"`
}

func GetUser(
	ctx context.Context,
	id uuid.UUID,
	dbtx db.DBTX,
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
		HashedPassword:  row.Password,
		IsAdmin:         false,
	}, nil
}

func NewUser(
	ctx context.Context,
	data NewUserPayload,
	dbtx db.DBTX,
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

	hashedPassword, err := HashAndPepperPassword(data.Password)
	if err != nil {
		return UserEntity{}, err
	}

	usr.HashedPassword = hashedPassword

	_, err = db.Stmts.InsertUser(ctx, dbtx, db.InsertUserParams{
		ID:        usr.ID,
		CreatedAt: pgtype.Timestamptz{Time: usr.CreatedAt, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: usr.UpdatedAt, Valid: true},
		Email:     usr.Email,
		Password:  usr.HashedPassword,
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
	data UpdateUserPayload,
	dbtx db.DBTX,
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
	if err := validate.Struct(data); err != nil {
		return errors.Join(ErrDomainValidation, err)
	}

	return q(ctx, data.ID, data.Password, data.UpdatedAt)
}

type UpdateUserEmailToVerifiedPayload struct {
	ID         uuid.UUID `validate:"required,uuid"`
	Email      string    `validate:"required,email"`
	VerifiedAt time.Time `validate:"required"`
}

func UpdateUserEmailToVerified(
	ctx context.Context,
	data UpdateUserEmailToVerifiedPayload,
	dbtx db.DBTX,
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
	data MakeUserAdminPayload,
	dbtx db.DBTX,
) (UserEntity, error) {
	if err := validate.Struct(data); err != nil {
		return UserEntity{}, errors.Join(ErrDomainValidation, err)
	}

	actor, err := GetUser(ctx, data.ActorID, dbtx)
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
