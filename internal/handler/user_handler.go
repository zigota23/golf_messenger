package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/yourusername/golf_messenger/internal/middleware"
	"github.com/yourusername/golf_messenger/internal/service"
	"github.com/yourusername/golf_messenger/pkg/response"
	"github.com/yourusername/golf_messenger/pkg/validator"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

type UpdateProfileRequest struct {
	FirstName string   `json:"first_name" validate:"omitempty,min=2,max=100"`
	LastName  string   `json:"last_name" validate:"omitempty,min=2,max=100"`
	Handicap  *float64 `json:"handicap" validate:"omitempty,gte=0,lte=54"`
	Phone     *string  `json:"phone" validate:"omitempty,max=20"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// GetMe godoc
// @Summary Get current user profile
// @Description Get the profile of the currently authenticated user
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=UserResponse} "User profile retrieved successfully"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 404 {object} response.Response "User not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/users/me [get]
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	user, err := h.userService.GetProfile(userID)
	if err != nil {
		if err.Error() == "user not found" {
			response.NotFound(w, err.Error())
			return
		}
		response.InternalServerError(w, "Failed to get user profile")
		return
	}

	userResp := UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Handicap:  user.Handicap,
		Phone:     user.Phone,
		AvatarURL: user.AvatarURL,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, userResp)
}

// UpdateMe godoc
// @Summary Update current user profile
// @Description Update the profile of the currently authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateProfileRequest true "Profile update details"
// @Success 200 {object} response.Response{data=UserResponse} "Profile updated successfully"
// @Failure 400 {object} response.Response "Bad request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 422 {object} response.Response "Validation error"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/users/me [put]
func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		errors := validator.FormatValidationErrors(err)
		response.UnprocessableEntity(w, "Validation failed", errors)
		return
	}

	user, err := h.userService.UpdateProfile(userID, req.FirstName, req.LastName, req.Handicap, req.Phone)
	if err != nil {
		if err.Error() == "user not found" {
			response.NotFound(w, err.Error())
			return
		}
		response.InternalServerError(w, "Failed to update profile")
		return
	}

	userResp := UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Handicap:  user.Handicap,
		Phone:     user.Phone,
		AvatarURL: user.AvatarURL,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, userResp)
}

// ChangePassword godoc
// @Summary Change user password
// @Description Change the password of the currently authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ChangePasswordRequest true "Password change details"
// @Success 200 {object} response.Response{data=map[string]string} "Password changed successfully"
// @Failure 400 {object} response.Response "Bad request"
// @Failure 401 {object} response.Response "Unauthorized or invalid old password"
// @Failure 422 {object} response.Response "Validation error"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/users/me/password [put]
func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		errors := validator.FormatValidationErrors(err)
		response.UnprocessableEntity(w, "Validation failed", errors)
		return
	}

	if err := h.userService.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		if err.Error() == "invalid old password" {
			response.Unauthorized(w, err.Error())
			return
		}
		if err.Error() == "user not found" {
			response.NotFound(w, err.Error())
			return
		}
		response.InternalServerError(w, "Failed to change password")
		return
	}

	response.Success(w, http.StatusOK, map[string]string{"message": "Password changed successfully"})
}

// UploadAvatar godoc
// @Summary Upload user avatar
// @Description Upload an avatar image for the currently authenticated user
// @Tags users
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param avatar formData file true "Avatar image file"
// @Success 200 {object} response.Response{data=UserResponse} "Avatar uploaded successfully"
// @Failure 400 {object} response.Response "Bad request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/users/me/avatar [post]
func (h *UserHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		response.BadRequest(w, "Failed to parse form data")
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		response.BadRequest(w, "Avatar file is required")
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/jpg" {
		response.BadRequest(w, "Only JPEG and PNG images are allowed")
		return
	}

	user, err := h.userService.UploadAvatar(r.Context(), userID, file, header.Filename, contentType)
	if err != nil {
		if err.Error() == "user not found" {
			response.NotFound(w, err.Error())
			return
		}
		response.InternalServerError(w, "Failed to upload avatar")
		return
	}

	userResp := UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Handicap:  user.Handicap,
		Phone:     user.Phone,
		AvatarURL: user.AvatarURL,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, userResp)
}

// DeleteAvatar godoc
// @Summary Delete user avatar
// @Description Delete the avatar of the currently authenticated user
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=UserResponse} "Avatar deleted successfully"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 404 {object} response.Response "User not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/users/me/avatar [delete]
func (h *UserHandler) DeleteAvatar(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	user, err := h.userService.DeleteAvatar(r.Context(), userID)
	if err != nil {
		if err.Error() == "user not found" {
			response.NotFound(w, err.Error())
			return
		}
		response.InternalServerError(w, "Failed to delete avatar")
		return
	}

	userResp := UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Handicap:  user.Handicap,
		Phone:     user.Phone,
		AvatarURL: user.AvatarURL,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, userResp)
}

// GetUserByID godoc
// @Summary Get user by ID
// @Description Get user profile by user ID
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} response.Response{data=UserResponse} "User profile retrieved successfully"
// @Failure 400 {object} response.Response "Invalid user ID"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 404 {object} response.Response "User not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	userID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		if err.Error() == "user not found" {
			response.NotFound(w, err.Error())
			return
		}
		response.InternalServerError(w, "Failed to get user")
		return
	}

	userResp := UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Handicap:  user.Handicap,
		Phone:     user.Phone,
		AvatarURL: user.AvatarURL,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, http.StatusOK, userResp)
}

// SearchUsers godoc
// @Summary Search users
// @Description Search users by name or email
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param q query string true "Search query"
// @Param limit query int false "Results limit" default(20)
// @Param offset query int false "Results offset" default(0)
// @Success 200 {object} response.Response{data=[]UserResponse} "Users retrieved successfully"
// @Failure 400 {object} response.Response "Bad request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/users [get]
func (h *UserHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		response.BadRequest(w, "Search query is required")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offsetStr := r.URL.Query().Get("offset")
	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	users, err := h.userService.SearchUsers(query, limit, offset)
	if err != nil {
		response.InternalServerError(w, "Failed to search users")
		return
	}

	userResponses := make([]UserResponse, 0, len(users))
	for _, user := range users {
		userResponses = append(userResponses, UserResponse{
			ID:        user.ID.String(),
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Handicap:  user.Handicap,
			Phone:     user.Phone,
			AvatarURL: user.AvatarURL,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	response.Success(w, http.StatusOK, userResponses)
}
