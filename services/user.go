package services

import (
	"context"

	"time"

	"github.com/go-playground/validator/v10"

	"github.com/MBvisti/grafto/entity"
	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/repository/database"
	"github.com/google/uuid"
)

type userDatabase interface {
	InsertUser(ctx context.Context, arg database.InsertUserParams) (database.User, error)
}

type NewUserData struct {
	ConfirmPassword string `validate:"required,gte=8"`
	Name            string `validate:"required,gte=2"`
	Mail            string `validate:"required,email"`
	Password        string `validate:"required,gte=8"`
}

func passwordMatchValidation(sl validator.StructLevel) {

	data := sl.Current().Interface().(NewUserData)

	if data.ConfirmPassword != data.Password {
		sl.ReportError(data.ConfirmPassword, "", "ConfirmPassword", "", "confirm password must match password")
	}
}

func NewUser(
	ctx context.Context, data NewUserData, db userDatabase, v *validator.Validate) (entity.User, error) {
	telemetry.Logger.Info("creating user")

	v.RegisterStructValidation(passwordMatchValidation, NewUserData{})

	if err := v.Struct(data); err != nil {
		return entity.User{}, err
	}

	hashedPassword, err := hashAndPepperPassword(data.Password)
	if err != nil {
		telemetry.Logger.Error("error", "value", err)
		return entity.User{}, err
	}

	user, err := db.InsertUser(ctx, database.InsertUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      data.Name,
		Mail:      data.Mail,
		Password:  hashedPassword,
	})
	if err != nil {
		return entity.User{}, err
	}

	return entity.User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Name:      user.Name,
		Mail:      user.Mail,
	}, nil
}
