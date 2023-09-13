package services

import (
	"context"
	"errors"
	"time"

	"github.com/MBvisti/grafto/entity"
	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/repository/database"
	"github.com/google/uuid"
	"github.com/gookit/validate"
)

type userDatabase interface {
	InsertUser(ctx context.Context, arg database.InsertUserParams) (database.User, error)
}

type NewUserData struct {
	Name            string `validate:"required|min_len:4" message:"required:{field} is required" label:"User Name"`
	Mail            string `validate:"required|email" message:"mail is invalid" label:"User Mail"`
	Password        string
	ConfirmPassword string
}

func NewUser(
	ctx context.Context, data NewUserData, db userDatabase) (entity.User, error) {
	telemetry.Logger.Info("creating user")

	v := validate.Struct(data)
	if err := v.ValidateE(); !err.Empty() {
		telemetry.Logger.Error("error creating new user", "error", err)
		return entity.User{}, errors.New("bad input")
	}

	userData := NewUserData{}
	if err := v.BindSafeData(&userData); err != nil {
		return entity.User{}, err
	}

	hashedPassword, err := hashAndPepperPassword(userData.Password)
	if err != nil {
		telemetry.Logger.Error("error", "value", err)
		return entity.User{}, err
	}

	user, err := db.InsertUser(ctx, database.InsertUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      userData.Name,
		Mail:      userData.Mail,
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
