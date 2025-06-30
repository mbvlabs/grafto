package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mbvlabs/grafto/models"
	"github.com/mbvlabs/grafto/psql"
)

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrCannotDeleteSelf      = errors.New("cannot delete your own account")
	ErrCannotRemoveLastAdmin = errors.New(
		"cannot remove admin privileges from last admin",
	)
)

type UserListItem struct {
	ID              uuid.UUID
	Email           string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	EmailVerifiedAt time.Time
	IsAdmin         bool
	IsVerified      bool
}

type UserListResponse struct {
	Users      []UserListItem
	TotalCount int
	Page       int
	PerPage    int
	TotalPages int
}

func GetAllUsers(
	ctx context.Context,
	database psql.Postgres,
	page int,
	perPage int,
) (UserListResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	rows, err := database.Pool.Query(
		ctx,
		"SELECT id, created_at, updated_at, email, email_verified_at, is_admin FROM users ORDER BY created_at DESC",
	)
	if err != nil {
		return UserListResponse{}, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []UserListItem
	for rows.Next() {
		var user UserListItem
		var emailVerifiedAt *time.Time
		err := rows.Scan(
			&user.ID,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.Email,
			&emailVerifiedAt,
			&user.IsAdmin,
		)
		if err != nil {
			return UserListResponse{}, fmt.Errorf(
				"failed to scan user: %w",
				err,
			)
		}

		if emailVerifiedAt != nil {
			user.EmailVerifiedAt = *emailVerifiedAt
			user.IsVerified = true
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return UserListResponse{}, fmt.Errorf("rows iteration error: %w", err)
	}

	totalCount := len(users)
	totalPages := (totalCount + perPage - 1) / perPage

	startIndex := (page - 1) * perPage
	endIndex := startIndex + perPage
	if endIndex > totalCount {
		endIndex = totalCount
	}

	var userList []UserListItem
	if startIndex < totalCount {
		userList = users[startIndex:endIndex]
	}

	return UserListResponse{
		Users:      userList,
		TotalCount: totalCount,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

func GetUserForEdit(
	ctx context.Context,
	database psql.Postgres,
	userID uuid.UUID,
) (models.User, error) {
	user, err := models.GetUser(ctx, database.Pool, userID)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

type UpdateUserDetailsPayload struct {
	UserID    uuid.UUID
	Email     string
	IsAdmin   bool
	ActorID   uuid.UUID
	UpdatedAt time.Time
}

func UpdateUserDetails(
	ctx context.Context,
	database psql.Postgres,
	payload UpdateUserDetailsPayload,
) (models.User, error) {
	actor, err := models.GetUser(ctx, database.Pool, payload.ActorID)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to get actor: %w", err)
	}

	if !actor.IsAdmin {
		return models.User{}, errors.New("only admins can update user details")
	}

	updatedUser, err := models.UpdateUser(
		ctx,
		database.Pool,
		models.UpdateUserPayload{
			ID:        payload.UserID,
			UpdatedAt: time.Now(),
			Email:     payload.Email,
		},
	)
	if err != nil {
		return models.User{}, err
	}

	return updatedUser, nil
}

type DeleteUserPayload struct {
	UserID  uuid.UUID
	ActorID uuid.UUID
}

func DeleteUser(
	ctx context.Context,
	database psql.Postgres,
	payload DeleteUserPayload,
) error {
	actor, err := models.GetUser(ctx, database.Pool, payload.ActorID)
	if err != nil {
		return fmt.Errorf("failed to get actor: %w", err)
	}

	if !actor.IsAdmin {
		return errors.New("only admins can delete users")
	}

	if payload.UserID == payload.ActorID {
		return ErrCannotDeleteSelf
	}

	userToDelete, err := models.GetUser(ctx, database.Pool, payload.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user to delete: %w", err)
	}

	if userToDelete.IsAdmin {
		adminCount, err := countAdminUsers(ctx, database)
		if err != nil {
			return fmt.Errorf("failed to check admin count: %w", err)
		}
		if adminCount <= 1 {
			return ErrCannotRemoveLastAdmin
		}
	}

	_, err = database.Pool.Exec(
		ctx,
		"DELETE FROM users WHERE id = $1",
		payload.UserID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

type ToggleUserAdminPayload struct {
	UserID  uuid.UUID
	ActorID uuid.UUID
}

func ToggleUserAdmin(
	ctx context.Context,
	database psql.Postgres,
	payload ToggleUserAdminPayload,
) (models.User, error) {
	actor, err := models.GetUser(ctx, database.Pool, payload.ActorID)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to get actor: %w", err)
	}

	if !actor.IsAdmin {
		return models.User{}, errors.New("only admins can toggle admin status")
	}

	if payload.UserID == payload.ActorID {
		return models.User{}, errors.New("cannot change your own admin status")
	}

	currentUser, err := models.GetUser(ctx, database.Pool, payload.UserID)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	newAdminStatus := !currentUser.IsAdmin

	if currentUser.IsAdmin && !newAdminStatus {
		adminCount, err := countAdminUsers(ctx, database)
		if err != nil {
			return models.User{}, fmt.Errorf(
				"failed to check admin count: %w",
				err,
			)
		}
		if adminCount <= 1 {
			return models.User{}, ErrCannotRemoveLastAdmin
		}
	}

	_, err = database.Pool.Exec(ctx,
		"UPDATE users SET is_admin = $2, updated_at = $3 WHERE id = $1",
		payload.UserID, newAdminStatus, time.Now())
	if err != nil {
		return models.User{}, fmt.Errorf(
			"failed to update user admin status: %w",
			err,
		)
	}

	// Get the updated user
	updatedUser, err := models.GetUser(ctx, database.Pool, payload.UserID)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to get updated user: %w", err)
	}

	return updatedUser, nil
}

func countAdminUsers(ctx context.Context, database psql.Postgres) (int, error) {
	var count int
	err := database.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE is_admin = true").
		Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count admin users: %w", err)
	}
	return count, nil
}
