package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/mbvlabs/grafto/psql"
	"github.com/mbvlabs/grafto/router/contexts"
	"github.com/mbvlabs/grafto/services"
	dashboardViews "github.com/mbvlabs/grafto/views/dashboards"
)

type Dashboard struct {
	db psql.Postgres
}

func newDashboard(db psql.Postgres) Dashboard {
	return Dashboard{db}
}

func (d Dashboard) Index(ctx echo.Context) error {
	return dashboardViews.Home().Render(renderArgs(ctx))
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
		if parsed, err := strconv.Atoi(pp); err == nil && parsed > 0 && parsed <= 100 {
			perPage = parsed
		}
	}

	userListResponse, err := services.GetAllUsers(ctx.Request().Context(), d.db, page, perPage)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, "Failed to load users")
	}

	return dashboardViews.UsersList(userListResponse, page, perPage).Render(renderArgs(ctx))
}

func (d Dashboard) EditUser(ctx echo.Context) error {
	userIDStr := ctx.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return ctx.String(http.StatusBadRequest, "Invalid user ID")
	}

	user, err := services.GetUserForEdit(ctx.Request().Context(), d.db, userID)
	if err != nil {
		return ctx.String(http.StatusNotFound, "User not found")
	}

	return dashboardViews.EditUser(user).Render(renderArgs(ctx))
}

func (d Dashboard) UpdateUser(ctx echo.Context) error {
	userIDStr := ctx.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return ctx.String(http.StatusBadRequest, "Invalid user ID")
	}

	email := ctx.FormValue("email")
	isAdminStr := ctx.FormValue("is_admin")
	isAdmin := isAdminStr == "on" || isAdminStr == "true"

	appCtx := contexts.ExtractApp(setAppCtx(ctx))
	currentUserID := appCtx.UserID

	updatePayload := services.UpdateUserDetailsPayload{
		UserID:    userID,
		Email:     email,
		IsAdmin:   isAdmin,
		ActorID:   currentUserID,
		UpdatedAt: time.Now(),
	}

	_, err = services.UpdateUserDetails(ctx.Request().Context(), d.db, updatePayload)
	if err != nil {
		if flashErr := addFlash(ctx, contexts.FlashError, fmt.Sprintf("Failed to update user: %v", err)); flashErr != nil {
			return flashErr
		}
		return ctx.Redirect(http.StatusSeeOther, fmt.Sprintf("/dashboard/users/%s/edit", userIDStr))
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
	currentUserID := appCtx.UserID

	deletePayload := services.DeleteUserPayload{
		UserID:  userID,
		ActorID: currentUserID,
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

	_, err = services.ToggleUserAdmin(ctx.Request().Context(), d.db, togglePayload)
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
