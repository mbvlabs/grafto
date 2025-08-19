package controllers

import (
	"errors"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/models"
	"github.com/mbvlabs/grafto/psql"
	"github.com/mbvlabs/grafto/router/cookies"
	"github.com/mbvlabs/grafto/services"
	"github.com/mbvlabs/grafto/views"
	"github.com/mbvlabs/grafto/views/fragments"
	"github.com/mbvlabs/grafto/views/sessions"
)

type Registrations struct {
	db          psql.Postgres
	emailClient services.EmailSender
}

func newRegistrations(
	db psql.Postgres,
	emailClient services.EmailSender,
) Registrations {
	return Registrations{db, emailClient}
}

func parseRegistrationErrors(err error) views.Errors {
	errs := views.Errors{}

	if errors.Is(err, services.ErrUserAlreadyExists) {
		errs[sessions.ErrEmailExists] = "This email is already registered"
		return errs
	}

	if errors.Is(err, models.ErrDomainValidation) {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			for _, validationErr := range validationErrs {
				switch validationErr.Field() {
				case "Email":
					if validationErr.Tag() == "email" {
						errs[sessions.EmailField] = "Please enter a valid email address"
					} else if validationErr.Tag() == "required" {
						errs[sessions.EmailField] = "Email is required"
					}
				case "Password":
					if validationErr.Tag() == "gte" {
						errs[sessions.PasswordField] = "Password must be at least 6 characters"
					} else if validationErr.Tag() == "required" {
						errs[sessions.PasswordField] = "Password is required"
					} else if validationErr.Tag() == "must match confirm password" {
						errs[sessions.ErrPasswordMismatch] = "Passwords do not match"
					}
				case "ConfirmPassword":
					if validationErr.Tag() == "gte" {
						errs[sessions.ConfirmPasswordField] = "Confirm password must be at least 6 characters"
					} else if validationErr.Tag() == "required" {
						errs[sessions.ConfirmPasswordField] = "Confirm password is required"
					} else if validationErr.Tag() == "must match password" {
						errs[sessions.ErrPasswordMismatch] = "Passwords do not match"
					}
				}
			}
		}
		return errs
	}

	errs[sessions.ErrInternalServer] = "An unexpected error occurred. Please try again."
	return errs
}

func (r Registrations) New(ctx echo.Context) error {
	return sessions.RegisterPage(sessions.RegisterFormProps{}).
		Render(renderArgs(ctx))
}

type StoreUserPayload struct {
	Email           string `form:"email"`
	Password        string `form:"password"`
	ConfirmPassword string `form:"confirm_password"`
}

func (r Registrations) Create(ctx echo.Context) error {
	var payload StoreUserPayload
	if err := ctx.Bind(&payload); err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	if err := services.RegisterUser(
		ctx.Request().Context(), r.db, r.emailClient, payload.Email, payload.Password, payload.ConfirmPassword); err != nil {
		slog.InfoContext(
			ctx.Request().Context(),
			"could not register user",
			"err",
			err,
		)

		userErrs := parseRegistrationErrors(err)
		return sessions.RegisterForm(sessions.RegisterFormProps{
			Errors:     userErrs,
			EmailValue: payload.Email,
		}).Render(renderArgs(ctx))
	}

	return fragments.VerifyCodeForm(fragments.VerifyCodeProps{
		CodeInvalid: false,
		Success:     false,
	}).Render(renderArgs(ctx))
}

type verificationCodePayload struct {
	Code string `form:"code"`
}

func (r Registrations) Update(ctx echo.Context) error {
	var payload verificationCodePayload
	if err := ctx.Bind(&payload); err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	user, err := services.ValidateUserEmail(
		ctx.Request().Context(),
		r.db,
		payload.Code,
	)
	if err != nil {
		return fragments.VerifyCodeForm(fragments.VerifyCodeProps{
			CodeInvalid: true,
			Success:     false,
		}).Render(renderArgs(ctx))
	}

	if err := cookies.CreateAuth(
		ctx, false, user); err != nil {
		return views.ErrorPage().Render(renderArgs(ctx))
	}

	return fragments.VerifyCodeForm(fragments.VerifyCodeProps{
		CodeInvalid: false,
		Success:     true,
	}).Render(renderArgs(ctx))
}
