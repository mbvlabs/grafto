package handlers

import (
	"time"

	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/models"
	"github.com/mbvlabs/grafto/psql"
	"github.com/mbvlabs/grafto/services"
	"github.com/mbvlabs/grafto/views"
	"github.com/mbvlabs/grafto/views/authentication"
)

type Registration struct {
	db       psql.Postgres
	emailSvc services.Email
}

func newRegistration(
	db psql.Postgres,
	emailSvc services.Email,
) Registration {
	return Registration{db, emailSvc}
}

func (r *Registration) CreateUser(ctx echo.Context) error {
	return authentication.RegisterPage(authentication.RegisterFormProps{
		CsrfToken: csrf.Token(ctx.Request()),
	}).Render(renderArgs(ctx))
}

type StoreUserPayload struct {
	Email           string `form:"email"`
	Password        string `form:"password"`
	ConfirmPassword string `form:"confirm_password"`
}

func (r *Registration) StoreUser(ctx echo.Context) error {
	var payload StoreUserPayload
	if err := ctx.Bind(&payload); err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	if _, err := models.NewUser(ctx.Request().Context(), models.NewUserPayload{
		Email:           payload.Email,
		Password:        payload.Password,
		ConfirmPassword: payload.ConfirmPassword,
	}, r.db.Pool); err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	// err := r.authSvc.RegisterUser(
	// 	ctx.Request().
	// 		Context(),
	// 	payload.UserName,
	// 	payload.Email,
	// 	payload.Password,
	// 	payload.ConfirmPassword,
	// )
	// if err != nil {
	// 	if errors.Is(err, services.ErrUnrecoverable) {
	// 		return views.ErrorPage().Render(renderArgs(ctx))
	// 	}
	//
	// 	if errors.Is(err, models.ErrDomainValidation) {
	// 		var validationErrors validator.ValidationErrors
	// 		if ok := errors.As(err, &validationErrors); !ok {
	// 			return views.ErrorPage().Render(renderArgs(ctx))
	// 		}
	//
	// 		fields := make(
	// 			map[string]components.InputFieldProps,
	// 			len(validationErrors),
	// 		)
	// 		for _, validationError := range validationErrors {
	// 			fields[validationError.StructField()] = components.InputFieldProps{
	// 				Value:     validationError.Value().(string),
	// 				ErrorMsgs: []string{validationError.Error()},
	// 			}
	// 		}
	//
	// 		props := authentication.RegisterFormProps{
	// 			SuccessRegister: false,
	// 			Fields:          fields,
	// 			CsrfToken:       csrf.Token(ctx.Request()),
	// 		}
	// 		return authentication.RegisterForm(props).
	// 			Render(renderArgs(ctx))
	// 	}
	// }

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

	token, err := models.GetToken(
		ctx.Request().Context(),
		payload.Token,
		r.db.Pool,
	)
	if err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	if !token.IsValid() || token.Meta.Scope != models.ScopeEmailVerification {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	user, err := models.GetUser(
		ctx.Request().Context(),
		token.Meta.ResourceID,
		r.db.Pool,
	)
	if err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	if err := models.UpdateUserEmailToVerified(
		ctx.Request().Context(),
		models.UpdateUserEmailToVerifiedPayload{
			ID:         user.ID,
			Email:      user.Email,
			VerifiedAt: time.Now(),
		},
		r.db.Pool,
	); err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	// TODO: create auth session
	return authentication.VerifyEmailPage(false).
		Render(renderArgs(ctx))
}
