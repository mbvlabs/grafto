package handlers

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/models"
	emails "github.com/mbvlabs/grafto/pkg/email_client"
	"github.com/mbvlabs/grafto/psql"
	"github.com/mbvlabs/grafto/services"
	"github.com/mbvlabs/grafto/views"
	"github.com/mbvlabs/grafto/views/authentication"
	"github.com/mbvlabs/grafto/views/components"
)

type Registration struct {
	authSvc     services.Auth
	db          psql.Postgres
	tknService  services.Token
	emailClient emails.EmailClient
}

func newRegistration(
	authSvc services.Auth,
	db psql.Postgres,
	tknService services.Token,
	emailClient emails.EmailClient,
) Registration {
	return Registration{authSvc, db, tknService, emailClient}
}

func (r *Registration) CreateUser(ctx echo.Context) error {
	return authentication.RegisterPage(authentication.RegisterFormProps{
		CsrfToken: csrf.Token(ctx.Request()),
	}).Render(renderArgs(ctx))
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
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	err := r.authSvc.RegisterUser(
		ctx.Request().
			Context(),
		payload.UserName,
		payload.Email,
		payload.Password,
		payload.ConfirmPassword,
	)
	if err != nil {
		if errors.Is(err, services.ErrUnrecoverable) {
			return views.ErrorPage().Render(renderArgs(ctx))
		}

		if errors.Is(err, models.ErrDomainValidation) {
			var validationErrors validator.ValidationErrors
			if ok := errors.As(err, &validationErrors); !ok {
				return views.ErrorPage().Render(renderArgs(ctx))
			}

			fields := make(
				map[string]components.InputFieldProps,
				len(validationErrors),
			)
			for _, validationError := range validationErrors {
				fields[validationError.StructField()] = components.InputFieldProps{
					Value:     validationError.Value().(string),
					ErrorMsgs: []string{validationError.Error()},
				}
			}

			props := authentication.RegisterFormProps{
				SuccessRegister: false,
				Fields:          fields,
				CsrfToken:       csrf.Token(ctx.Request()),
			}
			return authentication.RegisterForm(props).
				Render(renderArgs(ctx))
		}
	}

	props := authentication.RegisterFormProps{
		SuccessRegister: true,
		CsrfToken:       csrf.Token(ctx.Request()),
	}
	return authentication.RegisterForm(props).
		Render(renderArgs(ctx))
}

type verificationTokenPayload struct {
	Token string `query:"token"`
}

func (r *Registration) VerifyUserEmail(ctx echo.Context) error {
	var payload verificationTokenPayload
	if err := ctx.Bind(&payload); err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	if err := r.tknService.Validate(ctx.Request().Context(), payload.Token, services.ScopeEmailVerification); err != nil {
		if err := ctx.Bind(&payload); err != nil {
			ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
			ctx.Response().
				Writer.Header().
				Add("PreviousLocation", "/user/create")

			return views.ErrorPage().Render(renderArgs(ctx))
		}
	}

	// userID, err := r.tknService.GetAssociatedUserID(
	// 	ctx.Request().Context(),
	// 	payload.Token,
	// )
	// if err != nil {
	// 	if err := ctx.Bind(&payload); err != nil {
	// 		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
	// 		ctx.Response().
	// 			Writer.Header().
	// 			Add("PreviousLocation", "/user/create")
	//
	// 		return r.InternalError(ctx)
	// 	}
	// }
	//
	// user, err := r.db.QueryUserByID(ctx.Request().Context(), userID)
	// if err != nil {
	// 	return r.InternalError(ctx)
	// }

	// if err := r.userModel.VerifyEmail(ctx.Request().Context(), user.Email); err != nil {
	// 	return r.InternalError(ctx)
	// }
	//
	// _, err = r.authService.NewUserSession(
	// 	ctx.Request(),
	// 	ctx.Response(),
	// 	user.ID,
	// )
	// if err != nil {
	// 	return r.InternalError(ctx)
	// }

	return authentication.VerifyEmailPage(false).
		Render(renderArgs(ctx))
}
