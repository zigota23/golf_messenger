package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/yourusername/golf_messenger/internal/middleware"
	"github.com/yourusername/golf_messenger/internal/models"
	"github.com/yourusername/golf_messenger/internal/service"
	"github.com/yourusername/golf_messenger/pkg/response"
	"github.com/yourusername/golf_messenger/pkg/validator"
)

type TTRHandler struct {
	ttrService *service.TTRService
}

func NewTTRHandler(ttrService *service.TTRService) *TTRHandler {
	return &TTRHandler{ttrService: ttrService}
}

type CreateTTRRequest struct {
	CourseName     string `json:"course_name" validate:"required,min=2,max=255"`
	CourseLocation string `json:"course_location" validate:"omitempty,max=255"`
	TeeDate        string `json:"tee_date" validate:"required"`
	TeeTime        string `json:"tee_time" validate:"required"`
	MaxPlayers     int    `json:"max_players" validate:"required,min=1,max=8"`
	Notes          string `json:"notes" validate:"omitempty"`
}

type UpdateTTRRequest struct {
	CourseName     *string `json:"course_name" validate:"omitempty,min=2,max=255"`
	CourseLocation *string `json:"course_location" validate:"omitempty,max=255"`
	TeeDate        *string `json:"tee_date" validate:"omitempty"`
	TeeTime        *string `json:"tee_time" validate:"omitempty"`
	MaxPlayers     *int    `json:"max_players" validate:"omitempty,min=1,max=8"`
	Status         *string `json:"status" validate:"omitempty"`
	Notes          *string `json:"notes" validate:"omitempty"`
}

type AddCoCaptainRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
}

type UpdatePlayerStatusRequest struct {
	Status string `json:"status" validate:"required"`
}

type TTRResponse struct {
	ID              string              `json:"id"`
	CourseName      string              `json:"course_name"`
	CourseLocation  *string             `json:"course_location,omitempty"`
	TeeDate         string              `json:"tee_date"`
	TeeTime         string              `json:"tee_time"`
	MaxPlayers      int                 `json:"max_players"`
	CreatedByUserID string              `json:"created_by_user_id"`
	CaptainUserID   string              `json:"captain_user_id"`
	Status          string              `json:"status"`
	Notes           *string             `json:"notes,omitempty"`
	CreatedAt       string              `json:"created_at"`
	UpdatedAt       string              `json:"updated_at"`
	CreatedByUser   *UserResponse       `json:"created_by_user,omitempty"`
	CaptainUser     *UserResponse       `json:"captain_user,omitempty"`
	CoCaptains      []TTRCoCaptainResponse `json:"co_captains,omitempty"`
	Players         []TTRPlayerResponse `json:"players,omitempty"`
}

type TTRCoCaptainResponse struct {
	TTRID      string        `json:"ttr_id"`
	UserID     string        `json:"user_id"`
	AssignedAt string        `json:"assigned_at"`
	User       *UserResponse `json:"user,omitempty"`
}

type TTRPlayerResponse struct {
	TTRID    string        `json:"ttr_id"`
	UserID   string        `json:"user_id"`
	JoinedAt string        `json:"joined_at"`
	Status   string        `json:"status"`
	User     *UserResponse `json:"user,omitempty"`
}

// CreateTTR godoc
// @Summary Create new TTR
// @Description Create a new tee time reservation. The creator becomes the captain and is automatically added as the first player.
// @Tags ttrs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateTTRRequest true "TTR creation details"
// @Success 201 {object} response.Response{data=TTRResponse} "TTR created successfully"
// @Failure 400 {object} response.Response "Bad request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 422 {object} response.Response "Validation error"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/ttrs [post]
func (h *TTRHandler) CreateTTR(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	var req CreateTTRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		errors := validator.FormatValidationErrors(err)
		response.UnprocessableEntity(w, "Validation failed", errors)
		return
	}

	teeDate, err := time.Parse("2006-01-02", req.TeeDate)
	if err != nil {
		response.BadRequest(w, "Invalid tee_date format, expected YYYY-MM-DD")
		return
	}

	teeTime, err := time.Parse("15:04", req.TeeTime)
	if err != nil {
		response.BadRequest(w, "Invalid tee_time format, expected HH:MM")
		return
	}

	var courseLocation *string
	if req.CourseLocation != "" {
		courseLocation = &req.CourseLocation
	}

	var notes *string
	if req.Notes != "" {
		notes = &req.Notes
	}

	ttr, err := h.ttrService.CreateTTR(userID, req.CourseName, courseLocation, teeDate, teeTime, req.MaxPlayers, notes)
	if err != nil {
		response.InternalServerError(w, "Failed to create TTR")
		return
	}

	ttrResp := convertTTRToResponse(ttr)
	response.Success(w, http.StatusCreated, ttrResp)
}

// GetTTR godoc
// @Summary Get TTR by ID
// @Description Get detailed information about a specific TTR
// @Tags ttrs
// @Produce json
// @Security BearerAuth
// @Param id path string true "TTR ID (UUID)"
// @Success 200 {object} response.Response{data=TTRResponse} "TTR retrieved successfully"
// @Failure 400 {object} response.Response "Invalid TTR ID"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 404 {object} response.Response "TTR not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/ttrs/{id} [get]
func (h *TTRHandler) GetTTR(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	ttrID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid TTR ID")
		return
	}

	ttr, err := h.ttrService.GetTTR(ttrID)
	if err != nil {
		if err.Error() == "TTR not found" {
			response.NotFound(w, err.Error())
			return
		}
		response.InternalServerError(w, "Failed to get TTR")
		return
	}

	ttrResp := convertTTRToResponse(ttr)
	response.Success(w, http.StatusOK, ttrResp)
}

// UpdateTTR godoc
// @Summary Update TTR
// @Description Update TTR details. Only captain or co-captains can update.
// @Tags ttrs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "TTR ID (UUID)"
// @Param request body UpdateTTRRequest true "TTR update details"
// @Success 200 {object} response.Response{data=TTRResponse} "TTR updated successfully"
// @Failure 400 {object} response.Response "Bad request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Forbidden - not captain or co-captain"
// @Failure 404 {object} response.Response "TTR not found"
// @Failure 422 {object} response.Response "Validation error"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/ttrs/{id} [put]
func (h *TTRHandler) UpdateTTR(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	vars := mux.Vars(r)
	idStr := vars["id"]

	ttrID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid TTR ID")
		return
	}

	var req UpdateTTRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		errors := validator.FormatValidationErrors(err)
		response.UnprocessableEntity(w, "Validation failed", errors)
		return
	}

	var teeDate *time.Time
	if req.TeeDate != nil {
		parsed, err := time.Parse("2006-01-02", *req.TeeDate)
		if err != nil {
			response.BadRequest(w, "Invalid tee_date format, expected YYYY-MM-DD")
			return
		}
		teeDate = &parsed
	}

	var teeTime *time.Time
	if req.TeeTime != nil {
		parsed, err := time.Parse("15:04", *req.TeeTime)
		if err != nil {
			response.BadRequest(w, "Invalid tee_time format, expected HH:MM")
			return
		}
		teeTime = &parsed
	}

	ttr, err := h.ttrService.UpdateTTR(ttrID, userID, req.CourseName, req.CourseLocation, teeDate, teeTime, req.MaxPlayers, req.Status, req.Notes)
	if err != nil {
		if err.Error() == "TTR not found" {
			response.NotFound(w, err.Error())
			return
		}
		if err.Error() == "unauthorized: only captain or co-captain can update TTR" {
			response.Forbidden(w, err.Error())
			return
		}
		response.InternalServerError(w, "Failed to update TTR")
		return
	}

	ttrResp := convertTTRToResponse(ttr)
	response.Success(w, http.StatusOK, ttrResp)
}

// DeleteTTR godoc
// @Summary Delete TTR
// @Description Delete (cancel) a TTR. Only the captain can delete.
// @Tags ttrs
// @Produce json
// @Security BearerAuth
// @Param id path string true "TTR ID (UUID)"
// @Success 200 {object} response.Response{data=map[string]string} "TTR deleted successfully"
// @Failure 400 {object} response.Response "Invalid TTR ID"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Forbidden - not captain"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/ttrs/{id} [delete]
func (h *TTRHandler) DeleteTTR(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	vars := mux.Vars(r)
	idStr := vars["id"]

	ttrID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid TTR ID")
		return
	}

	if err := h.ttrService.DeleteTTR(ttrID, userID); err != nil {
		if err.Error() == "unauthorized: only captain can delete TTR" {
			response.Forbidden(w, err.Error())
			return
		}
		response.InternalServerError(w, "Failed to delete TTR")
		return
	}

	response.Success(w, http.StatusOK, map[string]string{"message": "TTR deleted successfully"})
}

// SearchTTRs godoc
// @Summary Search TTRs
// @Description Get a list of TTRs with optional filters
// @Tags ttrs
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Results limit" default(20)
// @Param offset query int false "Results offset" default(0)
// @Param status query string false "Filter by status (OPEN, CONFIRMED, CANCELLED, COMPLETED)"
// @Success 200 {object} response.Response{data=[]TTRResponse} "TTRs retrieved successfully"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/ttrs [get]
func (h *TTRHandler) SearchTTRs(w http.ResponseWriter, r *http.Request) {
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

	status := r.URL.Query().Get("status")

	ttrs, err := h.ttrService.SearchTTRs(limit, offset, status)
	if err != nil {
		response.InternalServerError(w, "Failed to search TTRs")
		return
	}

	ttrResponses := make([]TTRResponse, 0, len(ttrs))
	for _, ttr := range ttrs {
		ttrResponses = append(ttrResponses, convertTTRToResponse(ttr))
	}

	response.Success(w, http.StatusOK, ttrResponses)
}

// AddCoCaptain godoc
// @Summary Add co-captain to TTR
// @Description Add a user as co-captain. Only the captain can add co-captains.
// @Tags ttrs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "TTR ID (UUID)"
// @Param request body AddCoCaptainRequest true "Co-captain user ID"
// @Success 200 {object} response.Response{data=map[string]string} "Co-captain added successfully"
// @Failure 400 {object} response.Response "Bad request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Forbidden - not captain"
// @Failure 422 {object} response.Response "Validation error"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/ttrs/{id}/co-captains [post]
func (h *TTRHandler) AddCoCaptain(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	vars := mux.Vars(r)
	idStr := vars["id"]

	ttrID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid TTR ID")
		return
	}

	var req AddCoCaptainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		errors := validator.FormatValidationErrors(err)
		response.UnprocessableEntity(w, "Validation failed", errors)
		return
	}

	coCaptainUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}

	if err := h.ttrService.AddCoCaptain(ttrID, userID, coCaptainUserID); err != nil {
		if err.Error() == "unauthorized: only captain can add co-captains" {
			response.Forbidden(w, err.Error())
			return
		}
		if err.Error() == "co-captain user not found" {
			response.NotFound(w, err.Error())
			return
		}
		if err.Error() == "user is already a co-captain" {
			response.BadRequest(w, err.Error())
			return
		}
		response.InternalServerError(w, "Failed to add co-captain")
		return
	}

	response.Success(w, http.StatusOK, map[string]string{"message": "Co-captain added successfully"})
}

// RemoveCoCaptain godoc
// @Summary Remove co-captain from TTR
// @Description Remove a co-captain from the TTR. Only the captain can remove co-captains.
// @Tags ttrs
// @Produce json
// @Security BearerAuth
// @Param id path string true "TTR ID (UUID)"
// @Param userId path string true "User ID (UUID) of co-captain to remove"
// @Success 200 {object} response.Response{data=map[string]string} "Co-captain removed successfully"
// @Failure 400 {object} response.Response "Invalid ID"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Forbidden - not captain"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/ttrs/{id}/co-captains/{userId} [delete]
func (h *TTRHandler) RemoveCoCaptain(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	vars := mux.Vars(r)
	idStr := vars["id"]
	coCaptainIDStr := vars["userId"]

	ttrID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid TTR ID")
		return
	}

	coCaptainUserID, err := uuid.Parse(coCaptainIDStr)
	if err != nil {
		response.BadRequest(w, "Invalid user ID")
		return
	}

	if err := h.ttrService.RemoveCoCaptain(ttrID, userID, coCaptainUserID); err != nil {
		if err.Error() == "unauthorized: only captain can remove co-captains" {
			response.Forbidden(w, err.Error())
			return
		}
		response.InternalServerError(w, "Failed to remove co-captain")
		return
	}

	response.Success(w, http.StatusOK, map[string]string{"message": "Co-captain removed successfully"})
}

// JoinTTR godoc
// @Summary Join a TTR
// @Description Join a TTR as a player
// @Tags ttrs
// @Produce json
// @Security BearerAuth
// @Param id path string true "TTR ID (UUID)"
// @Success 200 {object} response.Response{data=map[string]string} "Joined TTR successfully"
// @Failure 400 {object} response.Response "Bad request or TTR is full"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 404 {object} response.Response "TTR not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/ttrs/{id}/join [post]
func (h *TTRHandler) JoinTTR(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	vars := mux.Vars(r)
	idStr := vars["id"]

	ttrID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid TTR ID")
		return
	}

	if err := h.ttrService.JoinTTR(ttrID, userID); err != nil {
		if err.Error() == "TTR not found" {
			response.NotFound(w, err.Error())
			return
		}
		if err.Error() == "TTR is full" || err.Error() == "user is already a player" {
			response.BadRequest(w, err.Error())
			return
		}
		response.InternalServerError(w, "Failed to join TTR")
		return
	}

	response.Success(w, http.StatusOK, map[string]string{"message": "Joined TTR successfully"})
}

// LeaveTTR godoc
// @Summary Leave a TTR
// @Description Leave a TTR. The captain cannot leave.
// @Tags ttrs
// @Produce json
// @Security BearerAuth
// @Param id path string true "TTR ID (UUID)"
// @Success 200 {object} response.Response{data=map[string]string} "Left TTR successfully"
// @Failure 400 {object} response.Response "Bad request or captain cannot leave"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 404 {object} response.Response "TTR not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/ttrs/{id}/leave [post]
func (h *TTRHandler) LeaveTTR(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	vars := mux.Vars(r)
	idStr := vars["id"]

	ttrID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid TTR ID")
		return
	}

	if err := h.ttrService.LeaveTTR(ttrID, userID); err != nil {
		if err.Error() == "TTR not found" {
			response.NotFound(w, err.Error())
			return
		}
		if err.Error() == "captain cannot leave TTR" {
			response.BadRequest(w, err.Error())
			return
		}
		response.InternalServerError(w, "Failed to leave TTR")
		return
	}

	response.Success(w, http.StatusOK, map[string]string{"message": "Left TTR successfully"})
}

// UpdatePlayerStatus godoc
// @Summary Update player status
// @Description Update a player's status in the TTR. Only captain or co-captains can update.
// @Tags ttrs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "TTR ID (UUID)"
// @Param userId path string true "Player User ID (UUID)"
// @Param request body UpdatePlayerStatusRequest true "New status"
// @Success 200 {object} response.Response{data=map[string]string} "Player status updated successfully"
// @Failure 400 {object} response.Response "Bad request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Forbidden - not captain or co-captain"
// @Failure 422 {object} response.Response "Validation error"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/ttrs/{id}/players/{userId} [put]
func (h *TTRHandler) UpdatePlayerStatus(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	vars := mux.Vars(r)
	idStr := vars["id"]
	playerIDStr := vars["userId"]

	ttrID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid TTR ID")
		return
	}

	playerUserID, err := uuid.Parse(playerIDStr)
	if err != nil {
		response.BadRequest(w, "Invalid player user ID")
		return
	}

	var req UpdatePlayerStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		errors := validator.FormatValidationErrors(err)
		response.UnprocessableEntity(w, "Validation failed", errors)
		return
	}

	if err := h.ttrService.UpdatePlayerStatus(ttrID, userID, playerUserID, req.Status); err != nil {
		if err.Error() == "unauthorized: only captain or co-captain can update player status" {
			response.Forbidden(w, err.Error())
			return
		}
		if err.Error() == "invalid player status" || err.Error() == "player not found in TTR" {
			response.BadRequest(w, err.Error())
			return
		}
		response.InternalServerError(w, "Failed to update player status")
		return
	}

	response.Success(w, http.StatusOK, map[string]string{"message": "Player status updated successfully"})
}

// GetPlayers godoc
// @Summary Get TTR players
// @Description Get all players for a specific TTR
// @Tags ttrs
// @Produce json
// @Security BearerAuth
// @Param id path string true "TTR ID (UUID)"
// @Success 200 {object} response.Response{data=[]TTRPlayerResponse} "Players retrieved successfully"
// @Failure 400 {object} response.Response "Invalid TTR ID"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/ttrs/{id}/players [get]
func (h *TTRHandler) GetPlayers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	ttrID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid TTR ID")
		return
	}

	players, err := h.ttrService.GetPlayers(ttrID)
	if err != nil {
		response.InternalServerError(w, "Failed to get players")
		return
	}

	playerResponses := make([]TTRPlayerResponse, 0, len(players))
	for _, player := range players {
		playerResp := TTRPlayerResponse{
			TTRID:    player.TTRID.String(),
			UserID:   player.UserID.String(),
			JoinedAt: player.JoinedAt.Format(time.RFC3339),
			Status:   player.Status,
		}
		if player.User != nil {
			userResp := convertUserToResponse(player.User)
			playerResp.User = &userResp
		}
		playerResponses = append(playerResponses, playerResp)
	}

	response.Success(w, http.StatusOK, playerResponses)
}

func convertTTRToResponse(ttr *models.TTR) TTRResponse {
	resp := TTRResponse{
		ID:              ttr.ID.String(),
		CourseName:      ttr.CourseName,
		CourseLocation:  ttr.CourseLocation,
		TeeDate:         ttr.TeeDate.Format("2006-01-02"),
		TeeTime:         ttr.TeeTime.Format("15:04"),
		MaxPlayers:      ttr.MaxPlayers,
		CreatedByUserID: ttr.CreatedByUserID.String(),
		CaptainUserID:   ttr.CaptainUserID.String(),
		Status:          ttr.Status,
		Notes:           ttr.Notes,
		CreatedAt:       ttr.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       ttr.UpdatedAt.Format(time.RFC3339),
	}

	if ttr.CreatedByUser != nil {
		userResp := convertUserToResponse(ttr.CreatedByUser)
		resp.CreatedByUser = &userResp
	}

	if ttr.CaptainUser != nil {
		userResp := convertUserToResponse(ttr.CaptainUser)
		resp.CaptainUser = &userResp
	}

	if ttr.CoCaptains != nil {
		resp.CoCaptains = make([]TTRCoCaptainResponse, 0, len(ttr.CoCaptains))
		for _, cc := range ttr.CoCaptains {
			ccResp := TTRCoCaptainResponse{
				TTRID:      cc.TTRID.String(),
				UserID:     cc.UserID.String(),
				AssignedAt: cc.AssignedAt.Format(time.RFC3339),
			}
			if cc.User != nil {
				userResp := convertUserToResponse(cc.User)
				ccResp.User = &userResp
			}
			resp.CoCaptains = append(resp.CoCaptains, ccResp)
		}
	}

	if ttr.Players != nil {
		resp.Players = make([]TTRPlayerResponse, 0, len(ttr.Players))
		for _, p := range ttr.Players {
			pResp := TTRPlayerResponse{
				TTRID:    p.TTRID.String(),
				UserID:   p.UserID.String(),
				JoinedAt: p.JoinedAt.Format(time.RFC3339),
				Status:   p.Status,
			}
			if p.User != nil {
				userResp := convertUserToResponse(p.User)
				pResp.User = &userResp
			}
			resp.Players = append(resp.Players, pResp)
		}
	}

	return resp
}

func convertUserToResponse(user *models.User) UserResponse {
	return UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Handicap:  user.Handicap,
		Phone:     user.Phone,
		AvatarURL: user.AvatarURL,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}
}
