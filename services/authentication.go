package services

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5"
	"github.com/mbv-labs/grafto/config"
	"github.com/mbv-labs/grafto/models"
	"github.com/mbv-labs/grafto/pkg/telemetry"
	"golang.org/x/crypto/bcrypt"
)

type authStorage interface {
	QueryUserByEmail(ctx context.Context, mail string) (models.User, error)
}

type Auth struct {
	storage     authStorage
	cookieStore *sessions.CookieStore
	cfg         config.Config
}

type UserSession struct {
	ID            uuid.UUID
	Authenticated bool
	IsAdmin       bool
}

func NewAuth(storage authStorage, cookieStore *sessions.CookieStore, cfg config.Config) Auth {
	return Auth{storage, cookieStore, cfg}
}

func (a Auth) HashAndPepperPassword(password string) (string, error) {
	passwordBytes := []byte(password + a.cfg.PasswordPepper)
	hashedBytes, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedBytes), nil
}

func (a Auth) ValidatePassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword(
		[]byte(hashedPassword),
		[]byte(password+a.cfg.PasswordPepper),
	)
}

func (a Auth) AuthenticateUser(
	ctx context.Context,
	email string,
	password string,
) error {
	user, err := a.storage.QueryUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotExist
		}

		telemetry.Logger.ErrorContext(ctx, "could not query user", "error", err)
		return err
	}

	if isVerified := user.IsVerified(); !isVerified {
		return ErrEmailNotValidated
	}

	hashedPw, err := a.HashAndPepperPassword(password)
	if err != nil {
		return err
	}

	if err := a.ValidatePassword(password, hashedPw); err != nil {
		return ErrPasswordNotMatch
	}

	return nil
}

func (a Auth) NewUserSession(
	req *http.Request,
	res http.ResponseWriter,
	userID uuid.UUID,
) (UserSession, error) {
	session, err := a.cookieStore.New(req, "ua")
	if err != nil {
		return UserSession{}, err
	}

	session.Options.HttpOnly = true
	session.Options.Domain = a.cfg.AppDomain
	session.Options.Secure = true
	session.Options.MaxAge = 86400

	session.Values["user_id"] = userID
	session.Values["authenticated"] = true
	session.Values["is_admin"] = false

	if err := session.Save(req, res); err != nil {
		return UserSession{}, err
	}

	return UserSession{
		ID:            userID,
		Authenticated: true,
		IsAdmin:       false,
	}, nil
}

func (a Auth) GetUserSession(req *http.Request) (UserSession, error) {
	session, err := a.cookieStore.Get(req, "ua")
	if err != nil {
		return UserSession{}, err
	}

	userID, ok := session.Values["user_id"].(uuid.UUID)
	if !ok {
		return UserSession{}, err
	}

	isAuthenticated, ok := session.Values["authenticated"].(bool)
	if !ok {
		return UserSession{}, err
	}

	isAdmin, ok := session.Values["is_admin"].(bool)
	if !ok {
		return UserSession{}, err
	}

	return UserSession{
		ID:            userID,
		Authenticated: isAuthenticated,
		IsAdmin:       isAdmin,
	}, nil
}
