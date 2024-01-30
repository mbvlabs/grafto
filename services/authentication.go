package services

import (
	"context"
	"encoding/gob"
	"errors"
	"net/http"
	"os"

	"github.com/MBvisti/grafto/entity"
	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/repository/database"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	// passwordPepper   = config.Cfg.GetPwdPepper()
	authSessionStore = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")), []byte(os.Getenv("SESSION_ENCRYPTION_KEY")))
)

func hashAndPepperPassword(password, passwordPepper string) (string, error) {
	passwordBytes := []byte(password + passwordPepper)
	hashedBytes, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedBytes), nil
}

type validatePasswordPayload struct {
	hashedpassword string
	password       string
}

func validatePassword(data validatePasswordPayload, passwordPepper string) error {
	return bcrypt.CompareHashAndPassword([]byte(data.hashedpassword), []byte(data.password+passwordPepper))
}

type AuthenticateUserPayload struct {
	Email    string
	Password string
}

func AuthenticateUser(ctx context.Context, data AuthenticateUserPayload, db userDatabase, passwordPepper string) (entity.User, error) {
	user, err := db.QueryUserByMail(ctx, data.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, ErrUserNotExist
		}

		telemetry.Logger.ErrorContext(ctx, "could not query user", "error", err)
		return entity.User{}, err
	}

	if verifiedAt := user.MailVerifiedAt; !verifiedAt.Valid {
		return entity.User{}, ErrEmailNotValidated
	}

	err = validatePassword(validatePasswordPayload{
		hashedpassword: user.Password,
		password:       data.Password,
	}, passwordPepper)
	if err != nil {
		return entity.User{}, ErrPasswordNotMatch
	}

	return entity.User{
		ID:        user.ID,
		CreatedAt: database.ConvertFromPGTimestamptzToTime(user.CreatedAt),
		UpdatedAt: database.ConvertFromPGTimestamptzToTime(user.UpdatedAt),
		Name:      user.Name,
		Mail:      user.Mail,
	}, nil
}

func CreateAuthenticatedSession(r *http.Request, w http.ResponseWriter, userID uuid.UUID) error {
	gob.Register(uuid.UUID{})
	session, err := authSessionStore.Get(r, "ua")
	if err != nil {
		return err
	}

	session.Options.HttpOnly = true
	session.Options.Domain = os.Getenv("APP_HOST")
	session.Options.Secure = true
	session.Options.MaxAge = 86400

	session.Values["user_id"] = userID
	session.Values["authenticated"] = true
	session.Values["is_admin"] = false

	return session.Save(r, w)
}

func IsAuthenticated(r *http.Request) (bool, uuid.UUID, error) {
	gob.Register(uuid.UUID{})
	session, err := authSessionStore.Get(r, "ua")
	if err != nil {
		return false, uuid.UUID{}, err
	}

	if session.Values["authenticated"] == nil {
		return false, uuid.UUID{}, err
	}

	return session.Values["authenticated"].(bool), session.Values["user_id"].(uuid.UUID), nil
}

func IsAdmin(r *http.Request) (bool, error) {
	gob.Register(uuid.UUID{})
	session, err := authSessionStore.Get(r, "ua")
	if err != nil {
		return false, err
	}

	return session.Values["is_admin"].(bool), nil
}
