package handlers

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/mbv-labs/grafto/models"
	"github.com/mbv-labs/grafto/pkg/mail/templates"
	"github.com/mbv-labs/grafto/pkg/queue"
	"github.com/mbv-labs/grafto/pkg/telemetry"
	"github.com/mbv-labs/grafto/pkg/tokens"
	"github.com/mbv-labs/grafto/pkg/validation"
	"github.com/mbv-labs/grafto/repository/psql/database"
	"github.com/mbv-labs/grafto/services"
	"github.com/mbv-labs/grafto/views"
	"github.com/mbv-labs/grafto/views/authentication"
)

type Registration struct {
	Base
	authService services.Auth
	userModel   models.UserService
	tknManager  tokens.Manager
}

func NewRegistration(
	authSvc services.Auth,
	base Base,
	userSvc models.UserService,
	tknManager tokens.Manager,
) Registration {
	return Registration{base, authSvc, userSvc, tknManager}
}

func (r *Registration) CreateUser(ctx echo.Context) error {
	return authentication.RegisterPage(authentication.RegisterFormProps{
		CsrfToken: csrf.Token(ctx.Request()),
	}, views.Head{}).Render(views.ExtractRenderDeps(ctx))
}

type StoreUserPayload struct {
	UserName        string `form:"user_name"`
	Mail            string `form:"email"`
	Password        string `form:"password"`
	ConfirmPassword string `form:"confirm_password"`
}

func (r *Registration) StoreUser(ctx echo.Context) error {
	var payload StoreUserPayload
	if err := ctx.Bind(&payload); err != nil {
		return authentication.RegisterResponse("An error occurred", "Please refresh the page an try again.", true).
			Render(views.ExtractRenderDeps(ctx))
	}

	t := time.Now()
	user, err := r.userModel.New(ctx.Request().Context(), models.CreateUserData{
		ID:              uuid.New(),
		CreatedAt:       t,
		UpdatedAt:       t,
		Name:            payload.UserName,
		Email:           payload.Mail,
		Password:        payload.Password,
		ConfirmPassword: payload.ConfirmPassword,
	})
	if err != nil && errors.Is(err, models.ErrUserAlreadyExists) { // TODO handle this better
		return authentication.RegisterResponse("An error occurred", "Please refresh the page an try again.", true).
			Render(views.ExtractRenderDeps(ctx))
	}
	if err != nil && errors.Is(err, models.ErrFailValidation) {
		var valiErr validation.ValidationErrors
		if ok := errors.As(err, &valiErr); !ok {
			return authentication.RegisterResponse("An error occurred", "Please refresh the page an try again.", true).
				Render(views.ExtractRenderDeps(ctx))
		}

		props := authentication.RegisterFormProps{
			NameInput: views.InputElementError{
				OldValue: payload.UserName,
			},
			EmailInput: views.InputElementError{
				OldValue: payload.Mail,
			},
			CsrfToken: csrf.Token(ctx.Request()),
		}

		for _, validationError := range valiErr {
			switch validationError.Field() {
			case "Name":
				props.NameInput.Invalid = true
				// TODO can be multiple errors and should be reflected in the UI
				props.NameInput.InvalidMsg = validationError.ErrorForHumans()
			case "MailRegistered":
				props.EmailInput.Invalid = true
				// TODO can be multiple errors and should be reflected in the UI
				props.EmailInput.InvalidMsg = validationError.ErrorForHumans()
			case "Password", "ConfirmPassword":
				props.PasswordInput = views.InputElementError{
					Invalid: true,
					// TODO can be multiple errors and should be reflected in the UI
					InvalidMsg: validationError.ErrorForHumans(),
				}
				props.ConfirmPassword = views.InputElementError{
					Invalid: true,
					// TODO can be multiple errors and should be reflected in the UI
					InvalidMsg: validationError.ErrorForHumans(),
				}
			}
		}

		return authentication.RegisterForm(props).Render(views.ExtractRenderDeps(ctx))
	}

	plainText, hashedToken, err := r.tknManager.GenerateToken()
	if err != nil {
		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)

		return authentication.RegisterResponse("An error occurred", "Please refresh the page an try again.", true).
			Render(views.ExtractRenderDeps(ctx))
	}

	activationToken := tokens.CreateActivationToken(plainText, hashedToken)

	if err := r.db.StoreToken(ctx.Request().Context(), database.StoreTokenParams{
		ID:        uuid.New(),
		CreatedAt: database.ConvertToPGTimestamptz(time.Now()),
		Hash:      activationToken.Hash,
		ExpiresAt: database.ConvertToPGTimestamptz(activationToken.GetExpirationTime()),
		Scope:     activationToken.GetScope(),
		UserID:    user.ID,
	}); err != nil {
		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)

		return authentication.RegisterResponse("An error occurred", "Please refresh the page an try again.", true).
			Render(views.ExtractRenderDeps(ctx))
	}

	userSignupMail := templates.UserSignupWelcomeMail{
		ConfirmationLink: fmt.Sprintf(
			"%s://%s/verify-email?token=%s",
			r.cfg.App.AppScheme,
			r.cfg.App.AppHost,
			activationToken.GetPlainText(),
		),
	}
	textVersion, err := userSignupMail.GenerateTextVersion()
	if err != nil {
		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)

		return authentication.RegisterResponse("An error occurred", "Please refresh the page an try again.", true).
			Render(views.ExtractRenderDeps(ctx))
	}
	htmlVersion, err := userSignupMail.GenerateHtmlVersion()
	if err != nil {
		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)

		return authentication.RegisterResponse("An error occurred", "Please refresh the page an try again.", true).
			Render(views.ExtractRenderDeps(ctx))
	}

	_, err = r.queueClient.Insert(ctx.Request().Context(), queue.EmailJobArgs{
		To:          user.Email,
		From:        r.cfg.App.DefaultSenderSignature,
		Subject:     "Thanks for signing up!",
		TextVersion: textVersion,
		HtmlVersion: htmlVersion,
	}, nil)
	if err != nil {
		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)

		return authentication.RegisterResponse("An error occurred", "Please refresh the page an try again.", true).
			Render(views.ExtractRenderDeps(ctx))
	}

	return authentication.RegisterResponse("You're now registered", "You should receive an email soon to validate your account.", false).
		Render(views.ExtractRenderDeps(ctx))
}

type verificationTokenPayload struct {
	Token string `query:"token"`
}

func (r *Registration) VerifyUserEmail(ctx echo.Context) error {
	var payload verificationTokenPayload
	if err := ctx.Bind(&payload); err != nil {
		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
		ctx.Response().Writer.Header().Add("PreviousLocation", "/user/create")

		return r.InternalError(ctx)
	}

	hashedToken, err := r.tknManager.Hash(payload.Token)
	if err != nil {
		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
		ctx.Response().Writer.Header().Add("PreviousLocation", "/login")

		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)
		return r.InternalError(ctx)
	}

	token, err := r.db.QueryTokenByHash(ctx.Request().Context(), hashedToken)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return authentication.VerifyEmailPage(true, views.Head{}).
				Render(views.ExtractRenderDeps(ctx))
		}

		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
		ctx.Response().Writer.Header().Add("PreviousLocation", "/login")

		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)
		return r.InternalError(ctx)
	}

	if database.ConvertFromPGTimestamptzToTime(token.ExpiresAt).Before(time.Now()) &&
		token.Scope != tokens.ScopeEmailVerification {
		return authentication.VerifyEmailPage(true, views.Head{}).
			Render(views.ExtractRenderDeps(ctx))
	}

	user, err := r.db.QueryUserByID(ctx.Request().Context(), token.UserID)
	if err != nil {
		return r.InternalError(ctx)
	}

	confirmTime := time.Now()
	updatedUser, err := r.db.UpdateUser(ctx.Request().Context(), database.UpdateUserParams{
		ID:             token.UserID,
		UpdatedAt:      database.ConvertToPGTimestamptz(confirmTime),
		Name:           user.Name,
		Mail:           user.Mail,
		Password:       user.Password,
		MailVerifiedAt: database.ConvertToPGTimestamptz(confirmTime),
	})
	if err != nil {
		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
		ctx.Response().Writer.Header().Add("PreviousLocation", "/login")

		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)
		return r.InternalError(ctx)
	}

	if err := r.db.DeleteToken(ctx.Request().Context(), token.ID); err != nil {
		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
		ctx.Response().Writer.Header().Add("PreviousLocation", "/login")

		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)
		return r.InternalError(ctx)
	}

	_, err = r.authService.NewUserSession(ctx.Request(), ctx.Response(), updatedUser.ID)
	if err != nil {
		return r.InternalError(ctx)
	}

	return authentication.VerifyEmailPage(false, views.Head{}).Render(views.ExtractRenderDeps(ctx))
}
