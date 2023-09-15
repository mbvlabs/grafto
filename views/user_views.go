package views

type RegisterUserData struct {
	NameInput       InputData
	EmailInput      InputData
	PasswordInput   InputData
	ConfirmPassword InputData
}

func (v Views) RegisterUser(ctx echo.Context, data RegisterUserData) error {
	return ctx.Render(http.StatusOK, "user/register", RenderOpts{
		Data: data,
	})
}

func (v Views) RegisteredUser(ctx echo.Context) error {
	return ctx.Render(http.StatusOK, "user/__registered", RenderOpts{
		Data: nil,
	})
}
