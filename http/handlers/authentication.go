package handlers

import (
	"log/slog"

	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
	emails "github.com/mbvlabs/grafto/pkg/email_client"
	"github.com/mbvlabs/grafto/psql"
	"github.com/mbvlabs/grafto/services"
	"github.com/mbvlabs/grafto/views"
	"github.com/mbvlabs/grafto/views/authentication"
)

type Authentication struct {
	authService services.Auth
	db          psql.Postgres
	tknService  services.Token
	emailClient emails.EmailClient
}

func newAuthentication(
	authSvc services.Auth,
	db psql.Postgres,
	tknService services.Token,
	emailClient emails.EmailClient,
) Authentication {
	return Authentication{authSvc, db, tknService, emailClient}
}

func (a *Authentication) CreateAuthenticatedSession(ctx echo.Context) error {
	return authentication.LoginPage(authentication.LoginPageProps{
		CsrfToken: csrf.Token(ctx.Request()),
	}).Render(extractRenderDeps(ctx))
}

type StoreAuthenticatedSessionPayload struct {
	Mail       string `form:"email"`
	Password   string `form:"password"`
	RememberMe string `form:"remember_me"`
}

func (a *Authentication) StoreAuthenticatedSession(ctx echo.Context) error {
	var payload StoreAuthenticatedSessionPayload
	if err := ctx.Bind(&payload); err != nil {
		slog.ErrorContext(
			ctx.Request().Context(),
			"could not parse UserLoginPayload",
			"error",
			err,
		)

		return authentication.LoginForm(csrf.Token(ctx.Request()), true, nil).
			Render(extractRenderDeps(ctx))
	}

	authedUser, err := a.authService.AuthenticateUser(
		ctx.Request().Context(),
		payload.Mail,
		payload.Password,
	)
	if err != nil {
		slog.ErrorContext(
			ctx.Request().Context(),
			"could not authenticate user",
			"error",
			err,
		)

		errors := make(views.Errors)

		switch err {
		case services.ErrPasswordNotMatch, services.ErrUserNotExist:
			errors[authentication.ErrAuthDetailsWrong] = "The email or password you entered is incorrect."
		case services.ErrEmailNotValidated:
			errors[authentication.ErrEmailNotValidated] = "Your email has not yet been verified."
		}

		return authentication.LoginForm(csrf.Token(ctx.Request()), false, errors).
			Render(extractRenderDeps(ctx))
	}

	if err := createAuthSession(ctx, true, authedUser); err != nil {
		internalError(ctx)
	}

	return authentication.LoginForm(csrf.Token(ctx.Request()), true, nil).
		Render(extractRenderDeps(ctx))
}

func (a *Authentication) CreatePasswordReset(ctx echo.Context) error {
	return authentication.ForgottenPasswordPage(csrf.Token(ctx.Request())).
		Render(extractRenderDeps(ctx))
}

type StorePasswordResetPayload struct {
	Email string `form:"email"`
}

func (a *Authentication) StorePasswordReset(ctx echo.Context) error {
	var payload StorePasswordResetPayload
	if err := ctx.Bind(&payload); err != nil {
		return authentication.ForgottenPasswordForm(authentication.ForgottenPasswordFormProps{
			CsrfToken:     csrf.Token(ctx.Request()),
			InternalError: true,
		}).
			Render(extractRenderDeps(ctx))
	}

	// user, err := a.db.QueryUserByEmail(ctx.Request().Context(), payload.Email)
	// if err != nil {
	// 	if errors.Is(err, pgx.ErrNoRows) {
	// 		return authentication.ForgottenPasswordForm(authentication.ForgottenPasswordFormProps{
	// 			CsrfToken:        csrf.Token(ctx.Request()),
	// 			NoAssociatedUser: true,
	// 		}).
	// 			Render(views.ExtractRenderDeps(ctx))
	// 	}
	//
	// 	return authentication.ForgottenPasswordForm(authentication.ForgottenPasswordFormProps{
	// 		CsrfToken:     csrf.Token(ctx.Request()),
	// 		InternalError: true,
	// 	}).
	// 		Render(views.ExtractRenderDeps(ctx))
	// }
	// resetToken, err := a.tknService.CreateResetPasswordToken(
	// 	ctx.Request().Context(),
	// 	user.ID,
	// )
	// if err != nil {
	// 	return err
	// }
	//
	// if err := a.emailClient.Send(ctx.Request().Context(), user.Email, re); err != nil {
	// 	return err
	// }

	return authentication.ForgottenPasswordForm(authentication.ForgottenPasswordFormProps{
		CsrfToken: csrf.Token(ctx.Request()),
		Success:   true,
	}).
		Render(extractRenderDeps(ctx))
}

type PasswordResetTokenPayload struct {
	Token string `query:"token"`
}

func (a *Authentication) CreateResetPassword(ctx echo.Context) error {
	var passwordResetToken PasswordResetTokenPayload
	if err := ctx.Bind(&passwordResetToken); err != nil {
		return authentication.ResetPasswordPage(false, true, csrf.Token(ctx.Request()), "").
			Render(extractRenderDeps(ctx))
	}

	return authentication.ResetPasswordPage(false, false, csrf.Token(ctx.Request()), passwordResetToken.Token).
		Render(extractRenderDeps(ctx))
}

type ResetPasswordPayload struct {
	Password        string `form:"password"`
	ConfirmPassword string `form:"confirm_password"`
	Token           string `form:"token"`
}

func (a *Authentication) StoreResetPassword(ctx echo.Context) error {
	var payload ResetPasswordPayload
	if err := ctx.Bind(&payload); err != nil {
		return authentication.ResetPasswordPage(false, true, "", "").
			Render(extractRenderDeps(ctx))
	}

	if err := a.tknService.Validate(ctx.Request().Context(), payload.Token, services.ScopeResetPassword); err != nil {
		return err
	}

	// userID, err := a.tknService.GetAssociatedUserID(
	// 	ctx.Request().Context(),
	// 	payload.Token,
	// )
	// if err != nil {
	// 	return authentication.ResetPasswordPage(false, true, "", "").
	// 		Render(views.ExtractRenderDeps(ctx))
	// }

	// err = a.userModel.ChangePassword(ctx.Request().Context(),
	// 	models.ChangeUserPasswordData{
	// 		ID:              userID,
	// 		UpdatedAt:       time.Now(),
	// 		Password:        payload.Password,
	// 		ConfirmPassword: payload.ConfirmPassword,
	// 	},
	// )
	// if err != nil && errors.Is(err, models.ErrFailValidation) {
	// 	var valiErrs validation.ValidationErrors
	// 	if ok := errors.As(err, &valiErrs); !ok {
	// 		return a.InternalError(ctx)
	// 	}
	//
	// 	props := authentication.ResetPasswordFormProps{
	// 		CsrfToken:  csrf.Token(ctx.Request()),
	// 		ResetToken: payload.Token,
	// 	}
	//
	// 	for _, validationError := range valiErrs {
	// 		switch validationError.GetFieldName() {
	// 		case "Password":
	// 			props.Errors[authentication.PasswordField] = validationError.GetHumanExplanations()[0]
	// 		case "ConfirmPassword":
	// 			props.Errors[authentication.PasswordField] = validationError.GetHumanExplanations()[0]
	// 		}
	// 	}
	//
	// 	return authentication.ResetPasswordForm(props).
	// 		Render(views.ExtractRenderDeps(ctx))
	// }
	// if err != nil {
	// 	return a.InternalError(ctx)
	// }

	if err := a.tknService.Delete(ctx.Request().Context(), payload.Token); err != nil {
		ctx.Response().Writer.Header().Add("HX-Redirect", "/500")
		ctx.Response().Writer.Header().Add("PreviousLocation", "/login")
		return internalError(ctx)
	}

	return authentication.ResetPasswordForm(authentication.ResetPasswordFormProps{}).
		Render(extractRenderDeps(ctx))
}
