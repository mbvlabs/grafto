package models

import (
	"context"
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mbv-labs/grafto/pkg/telemetry"
)

type userStorage interface {
	InsertUser(ctx context.Context, arg User, hashedPassword string) (User, error)
	QueryUserByEmail(ctx context.Context, mail string) (User, error)
	QueryUserByID(ctx context.Context, id uuid.UUID) (User, error)
	UpdateUser(ctx context.Context, arg User) (User, error)
}

type authService interface {
	HashAndPepperPassword(password string) (string, error)
}

type UserService struct {
	storage   userStorage
	authSvc   authService
	validator *validator.Validate
}

func NewUserService(storage userStorage, authSvc authService, v *validator.Validate) UserService {
	return UserService{storage, authSvc, v}
}

type NewUserValidation struct {
	ConfirmPassword string `validate:"required,gte=8"`
	Name            string `validate:"required,gte=2"`
	Mail            string `validate:"required,email"`
	MailRegistered  bool   `validate:"ne=true"`
	Password        string `validate:"required,gte=8"`
}

func PasswordMatchValidation(sl validator.StructLevel) {
	data := sl.Current().Interface().(NewUserValidation)

	if data.ConfirmPassword != data.Password {
		sl.ReportError(
			data.ConfirmPassword,
			"",
			"ConfirmPassword",
			"",
			"confirm password must match password",
		)
	}
}

func (us UserService) ByEmail(ctx context.Context, email string) (User, error) {
	user, err := us.storage.QueryUserByEmail(ctx, email)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (us UserService) New(
	ctx context.Context,
	data NewUserValidation,
) (User, error) {
	user, err := us.storage.QueryUserByEmail(ctx, data.Mail)
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		telemetry.Logger.Error("could not check if email exists", "error", err)
		return User{}, err
	}

	newUserData := NewUserValidation{
		ConfirmPassword: data.ConfirmPassword,
		Name:            data.Name,
		Mail:            data.Mail,
		MailRegistered:  user.Email != "",
		Password:        data.Password,
	}

	if err := us.validator.Struct(newUserData); err != nil {
		return User{}, err
	}

	hashedPassword, err := us.authSvc.HashAndPepperPassword(newUserData.Password)
	if err != nil {
		telemetry.Logger.Error("error hashing and peppering password", "error", err)
		return User{}, err
	}

	t := time.Now()
	newUser, err := us.storage.InsertUser(ctx, User{
		ID:        uuid.New(),
		CreatedAt: t,
		UpdatedAt: t,
		Name:      newUserData.Name,
		Email:     newUserData.Mail,
	}, hashedPassword)
	if err != nil {
		telemetry.Logger.Error("could not insert user", "error", err)
		return User{}, err
	}

	return newUser, nil
}

type UpdateUserValidation struct {
	ID   uuid.UUID `validate:"required"`
	Name string    `validate:"required,gte=2"`
	Mail string    `validate:"required,email"`
}

func (us UserService) UpdateUser(
	ctx context.Context,
	data UpdateUserValidation,
) (User, error) {
	validatedData := UpdateUserValidation{
		Name: data.Name,
		Mail: data.Mail,
	}

	if err := us.validator.Struct(validatedData); err != nil {
		return User{}, err
	}

	updatedUser, err := us.storage.UpdateUser(ctx, User{
		ID:        validatedData.ID,
		UpdatedAt: time.Now(),
		Name:      validatedData.Name,
		Email:     validatedData.Mail,
	})
	if err != nil {
		telemetry.Logger.Error("could not insert user", "error", err)
		return User{}, err
	}

	return updatedUser, nil
}
