package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type UserContext struct {
	echo.Context
	UserID          uuid.UUID
	IsAuthenticated bool
}

func (u *UserContext) GetID() uuid.UUID {
	return u.UserID
}

func (u *UserContext) GetAuthStatus() bool {
	return u.IsAuthenticated
}

type AdminContext struct {
	echo.Context
	isAdmin bool
}

func (a *AdminContext) GetAdminStatus() bool {
	return a.isAdmin
}
