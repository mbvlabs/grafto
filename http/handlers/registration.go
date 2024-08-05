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
	}).Render(views.ExtractRenderDeps(ctx))
}

type StoreUserPayload struct {
	UserName        string `form:"username"`
	Email           string `form:"email"`
	Password        string `form:"password"`
	ConfirmPassword string `form:"confirm_password"`
}

func (r *Registration) StoreUser(ctx echo.Context) error {
	var payload StoreUserPayload
	if err := ctx.Bind(&payload); err != nil {
		props := authentication.RegisterFormProps{
			InternalError: true,
			CsrfToken:     csrf.Token(ctx.Request()),
		}
		return authentication.RegisterForm(props).Render(views.ExtractRenderDeps(ctx))
	}

	t := time.Now()
	user, err := r.userModel.New(ctx.Request().Context(), models.CreateUserData{
		ID:              uuid.New(),
		CreatedAt:       t,
		UpdatedAt:       t,
		Name:            payload.UserName,
		Email:           payload.Email,
		Password:        payload.Password,
		ConfirmPassword: payload.ConfirmPassword,
	})
	if err != nil && errors.Is(err, models.ErrUserAlreadyExists) { // TODO handle this better
		props := authentication.RegisterFormProps{
			InternalError: true,
			CsrfToken:     csrf.Token(ctx.Request()),
		}
		return authentication.RegisterForm(props).Render(views.ExtractRenderDeps(ctx))
	}
	if err != nil && errors.Is(err, models.ErrFailValidation) {
		var valiErr validation.ValidationErrors
		if ok := errors.As(err, &valiErr); !ok {
			props := authentication.RegisterFormProps{
				InternalError: true,
				CsrfToken:     csrf.Token(ctx.Request()),
			}
			return authentication.RegisterForm(props).Render(views.ExtractRenderDeps(ctx))
		}

		props := authentication.RegisterFormProps{
			Fields: map[string]views.InputFieldProps{
				authentication.UsernameField: {
					Value: payload.UserName,
				},
				authentication.EmailField: {
					Value: payload.Email,
				},
				authentication.PasswordField: {},
			},
			CsrfToken: csrf.Token(ctx.Request()),
		}

		for _, validationError := range valiErr {
			switch validationError.GetFieldName() {
			case "Name":
				if entry, ok := props.Fields[authentication.UsernameField]; ok {
					entry.ErrorMsgs = validationError.GetHumanExplanations()
					props.Fields[authentication.UsernameField] = entry
				}
			case "Email":
				if entry, ok := props.Fields[authentication.EmailField]; ok {
					entry.ErrorMsgs = validationError.GetHumanExplanations()
					props.Fields[authentication.EmailField] = entry
				}
			case "Password":
				if entry, ok := props.Fields[authentication.PasswordField]; ok {
					entry.ErrorMsgs = validationError.GetHumanExplanations()
					props.Fields[authentication.PasswordField] = entry
				}
			}
		}

		return authentication.RegisterForm(props).Render(views.ExtractRenderDeps(ctx))
	}

	plainText, hashedToken, err := r.tknManager.GenerateToken()
	if err != nil {
		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)

		props := authentication.RegisterFormProps{
			InternalError: true,
			CsrfToken:     csrf.Token(ctx.Request()),
		}
		return authentication.RegisterForm(props).Render(views.ExtractRenderDeps(ctx))
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

		props := authentication.RegisterFormProps{
			InternalError: true,
			CsrfToken:     csrf.Token(ctx.Request()),
		}
		return authentication.RegisterForm(props).Render(views.ExtractRenderDeps(ctx))
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

		props := authentication.RegisterFormProps{
			InternalError: true,
			CsrfToken:     csrf.Token(ctx.Request()),
		}
		return authentication.RegisterForm(props).Render(views.ExtractRenderDeps(ctx))
	}
	htmlVersion, err := userSignupMail.GenerateHtmlVersion()
	if err != nil {
		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)

		props := authentication.RegisterFormProps{
			InternalError: true,
			CsrfToken:     csrf.Token(ctx.Request()),
		}
		return authentication.RegisterForm(props).Render(views.ExtractRenderDeps(ctx))
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

		props := authentication.RegisterFormProps{
			InternalError: true,
			CsrfToken:     csrf.Token(ctx.Request()),
		}
		return authentication.RegisterForm(props).Render(views.ExtractRenderDeps(ctx))
	}

	props := authentication.RegisterFormProps{
		SuccessRegister: true,
		CsrfToken:       csrf.Token(ctx.Request()),
	}
	return authentication.RegisterForm(props).Render(views.ExtractRenderDeps(ctx))
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
			return authentication.VerifyEmailPage(true).
				Render(views.ExtractRenderDeps(ctx))
		}

		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
		ctx.Response().Writer.Header().Add("PreviousLocation", "/login")

		telemetry.Logger.ErrorContext(ctx.Request().Context(), "could not query user", "error", err)
		return r.InternalError(ctx)
	}

	if database.ConvertFromPGTimestamptzToTime(token.ExpiresAt).Before(time.Now()) &&
		token.Scope != tokens.ScopeEmailVerification {
		return authentication.VerifyEmailPage(true).
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

	return authentication.VerifyEmailPage(false).Render(views.ExtractRenderDeps(ctx))
}
