package models

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/mbv-labs/grafto/pkg/telemetry"
	"github.com/mbv-labs/grafto/repository/database"
)

type userStorage interface {
	InsertUser(ctx context.Context, arg database.InsertUserParams) (database.User, error)
	DoesMailExists(ctx context.Context, mail string) (bool, error)
	QueryUserByMail(ctx context.Context, mail string) (database.User, error)
	QueryUser(ctx context.Context, id uuid.UUID) (database.User, error)
	UpdateUser(ctx context.Context, arg database.UpdateUserParams) (database.User, error)
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
	user, err := us.storage.QueryUserByMail(ctx, email)
	if err != nil {
		return User{}, err
	}

	return User{
		ID:             user.ID,
		CreatedAt:      database.ConvertFromPGTimestamptzToTime(user.CreatedAt),
		UpdatedAt:      database.ConvertFromPGTimestamptzToTime(user.UpdatedAt),
		Name:           user.Name,
		Mail:           user.Mail,
		MailVerifiedAt: database.ConvertFromPGTimestamptzToTime(user.MailVerifiedAt),
	}, nil
}

func (us UserService) New(
	ctx context.Context,
	data NewUserValidation,
) (User, error) {
	mailAlreadyRegistered, err := us.storage.DoesMailExists(ctx, data.Mail)
	if err != nil {
		telemetry.Logger.Error("could not check if email exists", "error", err)
		return User{}, err
	}

	// us.validator.RegisterStructValidation(passwordMatchValidation, newUserValidation{})

	newUserData := NewUserValidation{
		ConfirmPassword: data.ConfirmPassword,
		Name:            data.Name,
		Mail:            data.Mail,
		MailRegistered:  mailAlreadyRegistered,
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

	user, err := us.storage.InsertUser(ctx, database.InsertUserParams{
		ID:        uuid.New(),
		CreatedAt: database.ConvertToPGTimestamptz(time.Now()),
		UpdatedAt: database.ConvertToPGTimestamptz(time.Now()),
		Name:      newUserData.Name,
		Mail:      newUserData.Mail,
		Password:  hashedPassword,
	})
	if err != nil {
		telemetry.Logger.Error("could not insert user", "error", err)
		return User{}, err
	}

	return User{
		ID:        user.ID,
		CreatedAt: database.ConvertFromPGTimestamptzToTime(user.CreatedAt),
		UpdatedAt: database.ConvertFromPGTimestamptzToTime(user.UpdatedAt),
		Name:      user.Name,
		Mail:      user.Mail,
	}, nil
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
	// v.RegisterStructValidation(ResetPasswordMatchValidation, UpdateUserValidation{})

	validatedData := UpdateUserValidation{
		Name: data.Name,
		Mail: data.Mail,
	}

	if err := us.validator.Struct(validatedData); err != nil {
		return User{}, err
	}

	updatedUser, err := us.storage.UpdateUser(ctx, database.UpdateUserParams{
		UpdatedAt: database.ConvertToPGTimestamptz(time.Now()),
		Name:      data.Name,
		Mail:      data.Mail,
		ID:        data.ID,
	})
	if err != nil {
		telemetry.Logger.Error("could not insert user", "error", err)
		return User{}, err
	}

	return User{
		ID:        updatedUser.ID,
		CreatedAt: database.ConvertFromPGTimestamptzToTime(updatedUser.CreatedAt),
		UpdatedAt: database.ConvertFromPGTimestamptzToTime(updatedUser.UpdatedAt),
		Name:      updatedUser.Name,
		Mail:      updatedUser.Mail,
	}, nil
}
