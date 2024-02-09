package services

import (
	"context"

	"time"

	"github.com/MBvisti/grafto/entity"
	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/repository/database"
	"github.com/google/uuid"
)

type newUserValidation struct {
	ConfirmPassword string `validate:"required,gte=8"`
	Name            string `validate:"required,gte=2"`
	Mail            string `validate:"required,email"`
	MailRegistered  bool   `validate:"ne=true"`
	Password        string `validate:"required,gte=8"`
}

func (s *Services) NewUser(
	ctx context.Context, data entity.NewUser, passwordPepper string) (entity.User, error) {
	mailAlreadyRegistered, err := s.db.DoesMailExists(ctx, data.Mail)
	if err != nil {
		telemetry.Logger.Error("could not check if email exists", "error", err)
		return entity.User{}, err
	}

	newUserData := newUserValidation{
		ConfirmPassword: data.ConfirmPassword,
		Name:            data.Name,
		Mail:            data.Mail,
		MailRegistered:  mailAlreadyRegistered,
		Password:        data.Password,
	}

	if err := s.validator.Struct(newUserData); err != nil {
		return entity.User{}, err
	}

	hashedPassword, err := hashAndPepperPassword(newUserData.Password, passwordPepper)
	if err != nil {
		telemetry.Logger.Error("error hashing and peppering password", "error", err)
		return entity.User{}, err
	}

	user, err := s.db.InsertUser(ctx, database.InsertUserParams{
		ID:        uuid.New(),
		CreatedAt: database.ConvertToPGTimestamptz(time.Now()),
		UpdatedAt: database.ConvertToPGTimestamptz(time.Now()),
		Name:      newUserData.Name,
		Mail:      newUserData.Mail,
		Password:  hashedPassword,
	})
	if err != nil {
		telemetry.Logger.Error("could not insert user", "error", err)
		return entity.User{}, err
	}

	return entity.User{
		ID:        user.ID,
		CreatedAt: database.ConvertFromPGTimestamptzToTime(user.CreatedAt),
		UpdatedAt: database.ConvertFromPGTimestamptzToTime(user.UpdatedAt),
		Name:      user.Name,
		Mail:      user.Mail,
	}, nil
}

type updateUserValidation struct {
	ConfirmPassword string `validate:"required,gte=8"`
	Password        string `validate:"required,gte=8"`
	Name            string `validate:"required,gte=2"`
	Mail            string `validate:"required,email"`
}

func (s *Services) UpdateUser(
	ctx context.Context, data entity.UpdateUser, passwordPepper string) (entity.User, error) {

	validatedData := updateUserValidation{
		ConfirmPassword: data.ConfirmPassword,
		Password:        data.Password,
		Name:            data.Name,
		Mail:            data.Mail,
	}

	if err := s.validator.Struct(validatedData); err != nil {
		return entity.User{}, err
	}

	hashedPassword, err := hashAndPepperPassword(validatedData.Password, passwordPepper)
	if err != nil {
		telemetry.Logger.Error("error hashing and peppering password", "error", err)
		return entity.User{}, err
	}

	updatedUser, err := s.db.UpdateUser(ctx, database.UpdateUserParams{
		UpdatedAt: database.ConvertToPGTimestamptz(time.Now()),
		Name:      data.Name,
		Mail:      data.Mail,
		Password:  hashedPassword,
		ID:        data.ID,
	})
	if err != nil {
		telemetry.Logger.Error("could not insert user", "error", err)
		return entity.User{}, err
	}

	return entity.User{
		ID:        updatedUser.ID,
		CreatedAt: database.ConvertFromPGTimestamptzToTime(updatedUser.CreatedAt),
		UpdatedAt: database.ConvertFromPGTimestamptzToTime(updatedUser.UpdatedAt),
		Name:      updatedUser.Name,
		Mail:      updatedUser.Mail,
	}, nil
}
