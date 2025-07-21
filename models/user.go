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

type User struct {
	ID              uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Email           string
	EmailVerifiedAt time.Time
	IsAdmin         bool
	HashedPassword  string
}

func (u User) IsVerified() bool {
	return !u.EmailVerifiedAt.IsZero()
}

func (u User) ValidatePassword(providedPassword string) error {
	if t := subtle.ConstantTimeCompare(HashPassword(providedPassword), []byte(u.HashedPassword)); t == 1 {
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
		[]byte(config.Cfg.Auth.PasswordSalt),
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
) (User, error) {
	row, err := db.Stmts.QueryUserByEmail(ctx, dbtx, email)
	if err != nil {
		return User{}, err
	}

	return rowToUser(row), nil
}

func GetUser(
	ctx context.Context,
	dbtx db.DBTX,
	id uuid.UUID,
) (User, error) {
	row, err := db.Stmts.QueryUserByID(ctx, dbtx, id)
	if err != nil {
		return User{}, err
	}

	return rowToUser(row), nil
}

type NewUserPayload struct {
	Email    string `validate:"required,email"`
	Password PasswordPair
}

func NewUser(
	ctx context.Context,
	dbtx db.DBTX,
	data NewUserPayload,
) (User, error) {
	if err := validate.Struct(data); err != nil {
		return User{}, errors.Join(ErrDomainValidation, err)
	}

	usr := User{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Email:     data.Email,
	}
	hp := HashPassword(data.Password.Password)
	// usr.HashedPassword = string(hp)

	row, err := db.Stmts.InsertUser(ctx, dbtx, db.InsertUserParams{
		ID:        usr.ID,
		CreatedAt: pgtype.Timestamptz{Time: usr.CreatedAt, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: usr.UpdatedAt, Valid: true},
		Email:     usr.Email,
		Password:  hp,
	})
	if err != nil {
		return User{}, err
	}

	return rowToUser(row), nil
}

type UpdateUserPayload struct {
	ID              uuid.UUID `validate:"required,uuid"`
	Email           string
	EmailVerifiedAt time.Time
}

func UpdateUser(
	ctx context.Context,
	dbtx db.DBTX,
	data UpdateUserPayload,
) (User, error) {
	if err := validate.Struct(data); err != nil {
		return User{}, errors.Join(ErrDomainValidation, err)
	}

	currentRow, err := db.Stmts.QueryUserByID(ctx, dbtx, data.ID)
	if err != nil {
		return User{}, err
	}
	payload := db.UpdateUserParams{
		ID: data.ID,
		UpdatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		Email:           currentRow.Email,
		IsAdmin:         currentRow.IsAdmin,
		EmailVerifiedAt: currentRow.EmailVerifiedAt,
	}

	if data.Email != "" {
		payload.Email = data.Email
	}
	if !data.EmailVerifiedAt.IsZero() {
		payload.EmailVerifiedAt = pgtype.Timestamptz{
			Time:  data.EmailVerifiedAt,
			Valid: true,
		}
	}

	row, err := db.Stmts.UpdateUser(ctx, dbtx, payload)
	if err != nil {
		return User{}, err
	}

	return rowToUser(row), nil
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

type MakeUserAdminPayload struct {
	UserID    uuid.UUID `validate:"required,uuid"`
	UpdatedAt time.Time `validate:"required"`
}

func MakeUserAdmin(
	ctx context.Context,
	dbtx db.DBTX,
	data MakeUserAdminPayload,
) (User, error) {
	if err := validate.Struct(data); err != nil {
		return User{}, errors.Join(ErrDomainValidation, err)
	}

	currentRow, err := db.Stmts.QueryUserByID(ctx, dbtx, data.UserID)
	if err != nil {
		return User{}, err
	}
	payload := db.UpdateUserParams{
		ID: data.UserID,
		UpdatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		Email:           currentRow.Email,
		IsAdmin:         currentRow.IsAdmin,
		EmailVerifiedAt: currentRow.EmailVerifiedAt,
	}

	row, err := db.Stmts.UpdateUser(ctx, dbtx, payload)
	if err != nil {
		return User{}, err
	}

	return rowToUser(row), nil
}

type PaginatedUsers struct {
	Users      []User
	TotalCount int64
	Page       int32
	PageSize   int32
	TotalPages int32
}

func GetPaginatedUsers(
	ctx context.Context,
	dbtx db.DBTX,
	page int32,
	pageSize int32,
) (PaginatedUsers, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100 // Limit max page size
	}

	offset := (page - 1) * pageSize

	totalCount, err := db.Stmts.CountUsers(ctx, dbtx)
	if err != nil {
		return PaginatedUsers{}, err
	}

	rows, err := db.Stmts.QueryPaginatedUsers(
		ctx,
		dbtx,
		db.QueryPaginatedUsersParams{
			Limit:  pageSize,
			Offset: offset,
		},
	)
	if err != nil {
		return PaginatedUsers{}, err
	}

	users := make([]User, len(rows))
	for i, row := range rows {
		users[i] = rowToUser(row)
	}

	totalPages := int32((totalCount + int64(pageSize) - 1) / int64(pageSize))

	return PaginatedUsers{
		Users:      users,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func rowToUser(row db.User) User {
	return User{
		ID:              row.ID,
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
		Email:           row.Email,
		HashedPassword:  string(row.Password),
		EmailVerifiedAt: row.EmailVerifiedAt.Time,
		IsAdmin:         row.IsAdmin,
	}
}

func DeleteUser(
	ctx context.Context,
	dbtx db.DBTX,
	id uuid.UUID,
) error {
	return db.Stmts.DeleteUser(ctx, dbtx, id)
}
