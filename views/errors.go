package views

import "github.com/labstack/echo/v4"

type InternalServerErrData struct {
	FromLocation string
}

func InternalServerErr(ctx echo.Context, data InternalServerErrData) error {
	return nil
}
