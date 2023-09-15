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
	Name            string `validate:"required,gte=40"`
	Mail            string `validate:"required,email"`
	Password        string `validate:"required,gte=8"`
}

func NewUser(
	ctx context.Context, data NewUserData, db userDatabase) (entity.User, error) {
	telemetry.Logger.Info("creating user")

	validate := validator.New(validator.WithRequiredStructEnabled())

	if err := validate.Struct(data); err != nil {
		return entity.User{}, err
	}
	// if ok := v.Validate(); !ok {
	// 	telemetry.Logger.Error("error creating new user", "error", v.Errors.All())
	// 	return entity.User{}, errors.Wrap(ErrInvalidInput, v.Errors.Error())
	// }

	// userData := NewUserData{}
	// if err := v.BindSafeData(&userData); err != nil {
	// 	return entity.User{}, err
	// }

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
