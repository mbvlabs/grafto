package handlers

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/models"
	"github.com/mbvlabs/grafto/pkg/validation"
	"github.com/mbvlabs/grafto/services"
	"github.com/mbvlabs/grafto/views"
	"github.com/mbvlabs/grafto/views/authentication"
)

type Registration struct {
	Base
	authService  services.Auth
	userModel    models.UserService
	tknService   services.Token
	emailService services.Email
}

func NewRegistration(
	authSvc services.Auth,
	base Base,
	userSvc models.UserService,
	tknService services.Token,
	emailService services.Email,
) Registration {
	return Registration{base, authSvc, userSvc, tknService, emailService}
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

	emailActivationTkn, err := r.tknService.CreateUserEmailVerification(
		ctx.Request().Context(),
		user.ID,
	)
	if err != nil {
		props := authentication.RegisterFormProps{
			InternalError: true,
			CsrfToken:     csrf.Token(ctx.Request()),
		}
		return authentication.RegisterForm(props).Render(views.ExtractRenderDeps(ctx))
	}

	if err := r.emailService.SendUserSignupWelcome(ctx.Request().Context(), user.Email, emailActivationTkn, true); err != nil {
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

	if err := r.tknService.Validate(ctx.Request().Context(), payload.Token, services.ScopeEmailVerification); err != nil {
		if err := ctx.Bind(&payload); err != nil {
			ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
			ctx.Response().Writer.Header().Add("PreviousLocation", "/user/create")

			return r.InternalError(ctx)
		}
	}

	userID, err := r.tknService.GetAssociatedUserID(ctx.Request().Context(), payload.Token)
	if err != nil {
		if err := ctx.Bind(&payload); err != nil {
			ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
			ctx.Response().Writer.Header().Add("PreviousLocation", "/user/create")

			return r.InternalError(ctx)
		}
	}

	user, err := r.db.QueryUserByID(ctx.Request().Context(), userID)
	if err != nil {
		return r.InternalError(ctx)
	}

	if err := r.userModel.VerifyEmail(ctx.Request().Context(), user.Mail); err != nil {
		return r.InternalError(ctx)
	}

	_, err = r.authService.NewUserSession(ctx.Request(), ctx.Response(), user.ID)
	if err != nil {
		return r.InternalError(ctx)
	}

	return authentication.VerifyEmailPage(false).Render(views.ExtractRenderDeps(ctx))
}
