package handler

import (
	"encoding/json"
	"net/http"

	"github.com/yourusername/golf_messenger/internal/service"
	"github.com/yourusername/golf_messenger/pkg/response"
	"github.com/yourusername/golf_messenger/pkg/validator"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required,min=2,max=100"`
	LastName  string `json:"last_name" validate:"required,min=2,max=100"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    int64        `json:"expires_at"`
}

type UserResponse struct {
	ID        string   `json:"id"`
	Email     string   `json:"email"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Handicap  *float64 `json:"handicap,omitempty"`
	Phone     *string  `json:"phone,omitempty"`
	AvatarURL *string  `json:"avatar_url,omitempty"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration details"
// @Success 201 {object} response.Response{data=AuthResponse} "User registered successfully"
// @Failure 400 {object} response.Response "Bad request"
// @Failure 422 {object} response.Response "Validation error"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		errors := validator.FormatValidationErrors(err)
		response.UnprocessableEntity(w, "Validation failed", errors)
		return
	}

	user, tokenPair, err := h.authService.Register(req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		if err.Error() == "user with this email already exists" {
			response.Conflict(w, err.Error())
			return
		}
		response.InternalServerError(w, "Failed to register user")
		return
	}

	authResp := AuthResponse{
		User: UserResponse{
			ID:        user.ID.String(),
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Handicap:  user.Handicap,
			Phone:     user.Phone,
			AvatarURL: user.AvatarURL,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
	}

	response.Created(w, authResp)
}

// Login godoc
// @Summary Login user
// @Description Authenticate user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} response.Response{data=AuthResponse} "Login successful"
// @Failure 400 {object} response.Response "Bad request"
// @Failure 401 {object} response.Response "Invalid credentials"
// @Failure 422 {object} response.Response "Validation error"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		errors := validator.FormatValidationErrors(err)
		response.UnprocessableEntity(w, "Validation failed", errors)
		return
	}

	user, tokenPair, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		if err.Error() == "invalid email or password" {
			response.Unauthorized(w, err.Error())
			return
		}
		response.InternalServerError(w, "Failed to login")
		return
	}

	authResp := AuthResponse{
		User: UserResponse{
			ID:        user.ID.String(),
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Handicap:  user.Handicap,
			Phone:     user.Phone,
			AvatarURL: user.AvatarURL,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
	}

	response.Success(w, http.StatusOK, authResp)
}

// Refresh godoc
// @Summary Refresh access token
// @Description Get a new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh token"
// @Success 200 {object} response.Response{data=TokenResponse} "Token refreshed successfully"
// @Failure 400 {object} response.Response "Bad request"
// @Failure 401 {object} response.Response "Invalid refresh token"
// @Failure 422 {object} response.Response "Validation error"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		errors := validator.FormatValidationErrors(err)
		response.UnprocessableEntity(w, "Validation failed", errors)
		return
	}

	tokenPair, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		if err.Error() == "invalid refresh token" || err.Error() == "refresh token is invalid or expired" {
			response.Unauthorized(w, err.Error())
			return
		}
		response.InternalServerError(w, "Failed to refresh token")
		return
	}

	tokenResp := TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
	}

	response.Success(w, http.StatusOK, tokenResp)
}

// Logout godoc
// @Summary Logout user
// @Description Invalidate refresh token and logout user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh token"
// @Success 200 {object} response.Response{data=map[string]string} "Logout successful"
// @Failure 400 {object} response.Response "Bad request"
// @Failure 401 {object} response.Response "Invalid refresh token"
// @Failure 422 {object} response.Response "Validation error"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		errors := validator.FormatValidationErrors(err)
		response.UnprocessableEntity(w, "Validation failed", errors)
		return
	}

	if err := h.authService.Logout(req.RefreshToken); err != nil {
		if err.Error() == "invalid refresh token" {
			response.Unauthorized(w, err.Error())
			return
		}
		response.InternalServerError(w, "Failed to logout")
		return
	}

	response.Success(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}
