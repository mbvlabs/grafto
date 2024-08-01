package views

// type Error struct {
// 	Msg      string
// 	OldValue string
// }
//
// func (e Error) Exists() bool {
// 	return e.Msg != "" || e.OldValue != ""
// }

// type Errors map[string]Error

type Errors map[string]string

// import "github.com/labstack/echo/v4"
//
// type InternalServerErrData struct {
// 	FromLocation string
// }
//
// func InternalServerErr(ctx echo.Context, data InternalServerErrData) error {
// 	return nil
// }
