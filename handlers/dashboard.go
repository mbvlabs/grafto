package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/models"
	"github.com/mbvlabs/grafto/psql"
	"github.com/mbvlabs/grafto/router/reqmeta"
	"github.com/mbvlabs/grafto/views"
	"github.com/mbvlabs/grafto/views/dashboard"
)

type Dashboard struct {
	db psql.Postgres
}

func newDashboard(db psql.Postgres) Dashboard {
	return Dashboard{db}
}

func (d Dashboard) Index(c echo.Context) error {
	return dashboard.Home().Render(renderArgs(c))
}

func (d Dashboard) UsersList(c echo.Context) error {
	page := int64(1)
	if p := c.QueryParam("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = int64(parsed)
		}
	}

	perPage := int64(25)
	if pp := c.QueryParam("per_page"); pp != "" {
		if parsed, err := strconv.Atoi(pp); err == nil && parsed > 0 &&
			parsed <= 100 {
			perPage = int64(parsed)
		}
	}

	userListResponse, err := models.GetPaginatedUsers(
		c.Request().Context(),
		d.db.Pool,
		page,
		perPage,
	)
	if err != nil {
		return c.String(
			http.StatusInternalServerError,
			"Failed to load users",
		)
	}

	return dashboard.UsersList(userListResponse, page, perPage).
		Render(renderArgs(c))
}

func (d Dashboard) EditUser(c echo.Context) error {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid user ID")
	}

	user, err := models.GetUser(c.Request().Context(), d.db.Pool, userID)
	if err != nil {
		return c.String(http.StatusNotFound, "User not found")
	}

	return dashboard.EditUser(user).Render(renderArgs(c))
}

type UpdateUserPayload struct {
	Email         string `form:"email"`
	IsAdmin       string `form:"is_admin"`
	EmailVerified string `form:"is_verified"`
}

func (d Dashboard) UpdateUser(c echo.Context) error {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid user ID")
	}

	var payload UpdateUserPayload
	if err := c.Bind(&payload); err != nil {
		slog.ErrorContext(
			c.Request().Context(),
			"could not parse UpdateUserPayload",
			"error",
			err,
		)

		return views.ErrorPage().Render(renderArgs(c))
	}

	data := models.UpdateUserPayload{
		ID:    userID,
		Email: payload.Email,
	}
	if payload.EmailVerified == "on" {
		data.EmailVerifiedAt = time.Now()
	}

	_, err = models.UpdateUser(
		c.Request().Context(),
		d.db.Pool,
		data,
	)
	if err != nil {
		if flashErr := addFlash(c, reqmeta.FlashError, fmt.Sprintf("Failed to update user: %v", err)); flashErr != nil {
			return flashErr
		}
		return c.Redirect(
			http.StatusSeeOther,
			fmt.Sprintf("/dashboard/users/%s/edit", userID.String()),
		)
	}

	if flashErr := addFlash(c, reqmeta.FlashSuccess, "User updated successfully"); flashErr != nil {
		return flashErr
	}

	return c.Redirect(http.StatusSeeOther, "/dashboard/users")
}

func (d Dashboard) DeleteUser(c echo.Context) error {
	if err := adminOnlyAction(c, d.db.Pool); err != nil {
		return c.String(http.StatusUnauthorized, "User must be an admin")
	}

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid user ID")
	}

	err = models.DeleteUser(c.Request().Context(), d.db.Pool, userID)
	if err != nil {
		if flashErr := addFlash(c, reqmeta.FlashError, fmt.Sprintf("Failed to delete user: %v", err)); flashErr != nil {
			return flashErr
		}
	}

	if flashErr := addFlash(c, reqmeta.FlashSuccess, "User deleted successfully"); flashErr != nil {
		return flashErr
	}

	return c.Redirect(http.StatusSeeOther, "/dashboard/users")
}

func (d Dashboard) MakeUserAdmin(c echo.Context) error {
	if err := adminOnlyAction(c, d.db.Pool); err != nil {
		return c.String(http.StatusUnauthorized, "User must be an admin")
	}

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid user ID")
	}

	_, err = models.MakeUserAdmin(
		c.Request().Context(),
		d.db.Pool,
		models.MakeUserAdminPayload{
			UserID:    userID,
			UpdatedAt: time.Now(),
		},
	)
	if err != nil {
		if flashErr := addFlash(c, reqmeta.FlashError, fmt.Sprintf("Failed to make user admin: %v", err)); flashErr != nil {
			return flashErr
		}
	}

	if flashErr := addFlash(c, reqmeta.FlashSuccess, "Admin status updated successfully"); flashErr != nil {
		return flashErr
	}

	return c.Redirect(http.StatusSeeOther, "/dashboard/users")
}
