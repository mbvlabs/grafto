package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/psql"
	"github.com/mbvlabs/grafto/router/contexts"
	"github.com/mbvlabs/grafto/services"
	"github.com/mbvlabs/grafto/views"
	"github.com/mbvlabs/grafto/views/dashboard"
)

type Dashboard struct {
	db psql.Postgres
}

func newDashboard(db psql.Postgres) Dashboard {
	return Dashboard{db}
}

func (d Dashboard) Index(ctx echo.Context) error {
	return dashboard.Home().Render(renderArgs(ctx))
}

func (d Dashboard) UsersList(ctx echo.Context) error {
	page := 1
	if p := ctx.QueryParam("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	perPage := 25
	if pp := ctx.QueryParam("per_page"); pp != "" {
		if parsed, err := strconv.Atoi(pp); err == nil && parsed > 0 &&
			parsed <= 100 {
			perPage = parsed
		}
	}

	userListResponse, err := services.GetAllUsers(
		ctx.Request().Context(),
		d.db,
		page,
		perPage,
	)
	if err != nil {
		return ctx.String(
			http.StatusInternalServerError,
			"Failed to load users",
		)
	}

	return dashboard.UsersList(userListResponse, page, perPage).
		Render(renderArgs(ctx))
}

func (d Dashboard) EditUser(ctx echo.Context) error {
	userIDParam := ctx.Param("id")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		return ctx.String(http.StatusBadRequest, "Invalid user ID")
	}

	user, err := services.GetUserForEdit(ctx.Request().Context(), d.db, userID)
	if err != nil {
		return ctx.String(http.StatusNotFound, "User not found")
	}

	return dashboard.EditUser(user).Render(renderArgs(ctx))
}

type UpdateUserPayload struct {
	Email         string `form:"email"`
	IsAdmin       string `form:"is_admin"`
	EmailVerified string `form:"is_verified"`
}

func (d Dashboard) UpdateUser(ctx echo.Context) error {
	userIDParam := ctx.Param("id")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		return ctx.String(http.StatusBadRequest, "Invalid user ID")
	}

	var payload UpdateUserPayload
	if err := ctx.Bind(&payload); err != nil {
		slog.ErrorContext(
			ctx.Request().Context(),
			"could not parse UpdateUserPayload",
			"error",
			err,
		)

		return views.ErrorPage().Render(renderArgs(ctx))
	}

	appCtx := contexts.ExtractApp(setAppCtx(ctx))
	_, err = services.UpdateUserDetails(
		ctx.Request().Context(),
		d.db,
		services.UpdateUserDetailsPayload{
			UserID:    userID,
			Email:     payload.Email,
			IsAdmin:   payload.IsAdmin == "on",
			ActorID:   appCtx.UserID,
			UpdatedAt: time.Now(),
		},
	)
	if err != nil {
		if flashErr := addFlash(ctx, contexts.FlashError, fmt.Sprintf("Failed to update user: %v", err)); flashErr != nil {
			return flashErr
		}
		return ctx.Redirect(
			http.StatusSeeOther,
			fmt.Sprintf("/dashboard/users/%s/edit", userIDParam),
		)
	}

	if flashErr := addFlash(ctx, contexts.FlashSuccess, "User updated successfully"); flashErr != nil {
		return flashErr
	}
	return ctx.Redirect(http.StatusSeeOther, "/dashboard/users")
}

func (d Dashboard) DeleteUser(ctx echo.Context) error {
	userIDStr := ctx.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return ctx.String(http.StatusBadRequest, "Invalid user ID")
	}

	appCtx := contexts.ExtractApp(setAppCtx(ctx))

	deletePayload := services.DeleteUserPayload{
		UserID:  userID,
		ActorID: appCtx.UserID,
	}

	err = services.DeleteUser(ctx.Request().Context(), d.db, deletePayload)
	if err != nil {
		if flashErr := addFlash(ctx, contexts.FlashError, fmt.Sprintf("Failed to delete user: %v", err)); flashErr != nil {
			return flashErr
		}
	} else {
		if flashErr := addFlash(ctx, contexts.FlashSuccess, "User deleted successfully"); flashErr != nil {
			return flashErr
		}
	}

	return ctx.Redirect(http.StatusSeeOther, "/dashboard/users")
}

func (d Dashboard) ToggleUserAdmin(ctx echo.Context) error {
	userIDStr := ctx.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return ctx.String(http.StatusBadRequest, "Invalid user ID")
	}

	appCtx := contexts.ExtractApp(setAppCtx(ctx))
	currentUserID := appCtx.UserID

	togglePayload := services.ToggleUserAdminPayload{
		UserID:  userID,
		ActorID: currentUserID,
	}

	_, err = services.ToggleUserAdmin(
		ctx.Request().Context(),
		d.db,
		togglePayload,
	)
	if err != nil {
		if flashErr := addFlash(ctx, contexts.FlashError, fmt.Sprintf("Failed to toggle admin status: %v", err)); flashErr != nil {
			return flashErr
		}
	} else {
		if flashErr := addFlash(ctx, contexts.FlashSuccess, "Admin status updated successfully"); flashErr != nil {
			return flashErr
		}
	}

	return ctx.Redirect(http.StatusSeeOther, "/dashboard/users")
}
