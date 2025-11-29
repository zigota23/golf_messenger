package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/yourusername/golf_messenger/internal/middleware"
	"github.com/yourusername/golf_messenger/internal/models"
	"github.com/yourusername/golf_messenger/internal/service"
	"github.com/yourusername/golf_messenger/pkg/response"
	"github.com/yourusername/golf_messenger/pkg/validator"
)

type InvitationHandler struct {
	invitationService *service.InvitationService
}

func NewInvitationHandler(invitationService *service.InvitationService) *InvitationHandler {
	return &InvitationHandler{invitationService: invitationService}
}

type CreateInvitationRequest struct {
	TTRID         string `json:"ttr_id" validate:"required,uuid"`
	InviteeUserID string `json:"invitee_user_id" validate:"required,uuid"`
	Message       string `json:"message" validate:"omitempty"`
}

type RespondToInvitationRequest struct {
	Status string `json:"status" validate:"required"`
}

type InvitationResponse struct {
	ID            string        `json:"id"`
	TTRID         string        `json:"ttr_id"`
	InviterUserID string        `json:"inviter_user_id"`
	InviteeUserID string        `json:"invitee_user_id"`
	Status        string        `json:"status"`
	Message       *string       `json:"message,omitempty"`
	CreatedAt     string        `json:"created_at"`
	RespondedAt   *string       `json:"responded_at,omitempty"`
	TTR           *TTRResponse  `json:"ttr,omitempty"`
	InviterUser   *UserResponse `json:"inviter_user,omitempty"`
	InviteeUser   *UserResponse `json:"invitee_user,omitempty"`
}

// CreateInvitation godoc
// @Summary Create invitation
// @Description Send an invitation to a user to join a TTR. Only captain or co-captains can send invitations.
// @Tags invitations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateInvitationRequest true "Invitation details"
// @Success 201 {object} response.Response{data=InvitationResponse} "Invitation created successfully"
// @Failure 400 {object} response.Response "Bad request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Forbidden - not captain or co-captain"
// @Failure 404 {object} response.Response "TTR or user not found"
// @Failure 422 {object} response.Response "Validation error"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/invitations [post]
func (h *InvitationHandler) CreateInvitation(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	var req CreateInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		errors := validator.FormatValidationErrors(err)
		response.UnprocessableEntity(w, "Validation failed", errors)
		return
	}

	ttrID, err := uuid.Parse(req.TTRID)
	if err != nil {
		response.BadRequest(w, "Invalid TTR ID")
		return
	}

	inviteeUserID, err := uuid.Parse(req.InviteeUserID)
	if err != nil {
		response.BadRequest(w, "Invalid invitee user ID")
		return
	}

	var message *string
	if req.Message != "" {
		message = &req.Message
	}

	invitation, err := h.invitationService.CreateInvitation(ttrID, userID, inviteeUserID, message)
	if err != nil {
		if err.Error() == "TTR not found" || err.Error() == "invitee user not found" {
			response.NotFound(w, err.Error())
			return
		}
		if err.Error() == "unauthorized: only captain or co-captain can send invitations" {
			response.Forbidden(w, err.Error())
			return
		}
		if err.Error() == "TTR is full" || err.Error() == "invitee is already a player in this TTR" || err.Error() == "pending invitation already exists for this user" {
			response.BadRequest(w, err.Error())
			return
		}
		response.InternalServerError(w, "Failed to create invitation")
		return
	}

	invitationResp := convertInvitationToResponse(invitation)
	response.Success(w, http.StatusCreated, invitationResp)
}

// RespondToInvitation godoc
// @Summary Respond to invitation
// @Description Respond to a received invitation with YES, NO, or MAYBE
// @Tags invitations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Invitation ID (UUID)"
// @Param request body RespondToInvitationRequest true "Response status"
// @Success 200 {object} response.Response{data=InvitationResponse} "Response recorded successfully"
// @Failure 400 {object} response.Response "Bad request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 404 {object} response.Response "Invitation not found"
// @Failure 422 {object} response.Response "Validation error"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/invitations/{id}/respond [put]
func (h *InvitationHandler) RespondToInvitation(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	vars := mux.Vars(r)
	idStr := vars["id"]

	invitationID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid invitation ID")
		return
	}

	var req RespondToInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		errors := validator.FormatValidationErrors(err)
		response.UnprocessableEntity(w, "Validation failed", errors)
		return
	}

	invitation, err := h.invitationService.RespondToInvitation(invitationID, userID, req.Status)
	if err != nil {
		if err.Error() == "invitation not found" || err.Error() == "TTR not found" {
			response.NotFound(w, err.Error())
			return
		}
		if err.Error() == "unauthorized: you can only respond to your own invitations" {
			response.Forbidden(w, err.Error())
			return
		}
		if err.Error() == "invalid invitation status" || err.Error() == "invitation has already been responded to" || err.Error() == "TTR is full, cannot accept invitation" {
			response.BadRequest(w, err.Error())
			return
		}
		response.InternalServerError(w, "Failed to respond to invitation")
		return
	}

	invitationResp := convertInvitationToResponse(invitation)
	response.Success(w, http.StatusOK, invitationResp)
}

// GetInvitation godoc
// @Summary Get invitation by ID
// @Description Get detailed information about a specific invitation
// @Tags invitations
// @Produce json
// @Security BearerAuth
// @Param id path string true "Invitation ID (UUID)"
// @Success 200 {object} response.Response{data=InvitationResponse} "Invitation retrieved successfully"
// @Failure 400 {object} response.Response "Invalid invitation ID"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 404 {object} response.Response "Invitation not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/invitations/{id} [get]
func (h *InvitationHandler) GetInvitation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	invitationID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid invitation ID")
		return
	}

	invitation, err := h.invitationService.GetInvitation(invitationID)
	if err != nil {
		if err.Error() == "invitation not found" {
			response.NotFound(w, err.Error())
			return
		}
		response.InternalServerError(w, "Failed to get invitation")
		return
	}

	invitationResp := convertInvitationToResponse(invitation)
	response.Success(w, http.StatusOK, invitationResp)
}

// GetMyInvitations godoc
// @Summary Get my invitations
// @Description Get invitations received or sent by the authenticated user
// @Tags invitations
// @Produce json
// @Security BearerAuth
// @Param type query string false "Filter by type: 'received' or 'sent'" default(received)
// @Success 200 {object} response.Response{data=[]InvitationResponse} "Invitations retrieved successfully"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/invitations/me [get]
func (h *InvitationHandler) GetMyInvitations(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	invitationType := r.URL.Query().Get("type")
	received := true
	if invitationType == "sent" {
		received = false
	}

	invitations, err := h.invitationService.GetUserInvitations(userID, received)
	if err != nil {
		response.InternalServerError(w, "Failed to get invitations")
		return
	}

	invitationResponses := make([]InvitationResponse, 0, len(invitations))
	for _, invitation := range invitations {
		invitationResponses = append(invitationResponses, convertInvitationToResponse(invitation))
	}

	response.Success(w, http.StatusOK, invitationResponses)
}

// CancelInvitation godoc
// @Summary Cancel invitation
// @Description Cancel a pending invitation. Only the inviter can cancel.
// @Tags invitations
// @Produce json
// @Security BearerAuth
// @Param id path string true "Invitation ID (UUID)"
// @Success 200 {object} response.Response{data=map[string]string} "Invitation canceled successfully"
// @Failure 400 {object} response.Response "Bad request or invitation already responded"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 403 {object} response.Response "Forbidden - not the inviter"
// @Failure 404 {object} response.Response "Invitation not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /api/v1/invitations/{id} [delete]
func (h *InvitationHandler) CancelInvitation(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	vars := mux.Vars(r)
	idStr := vars["id"]

	invitationID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "Invalid invitation ID")
		return
	}

	if err := h.invitationService.CancelInvitation(invitationID, userID); err != nil {
		if err.Error() == "invitation not found" {
			response.NotFound(w, err.Error())
			return
		}
		if err.Error() == "unauthorized: only the inviter can cancel the invitation" {
			response.Forbidden(w, err.Error())
			return
		}
		if err.Error() == "only pending invitations can be canceled" {
			response.BadRequest(w, err.Error())
			return
		}
		response.InternalServerError(w, "Failed to cancel invitation")
		return
	}

	response.Success(w, http.StatusOK, map[string]string{"message": "Invitation canceled successfully"})
}

func convertInvitationToResponse(invitation *models.Invitation) InvitationResponse {
	resp := InvitationResponse{
		ID:            invitation.ID.String(),
		TTRID:         invitation.TTRID.String(),
		InviterUserID: invitation.InviterUserID.String(),
		InviteeUserID: invitation.InviteeUserID.String(),
		Status:        invitation.Status,
		Message:       invitation.Message,
		CreatedAt:     invitation.CreatedAt.Format(time.RFC3339),
	}

	if invitation.RespondedAt != nil {
		respondedAt := invitation.RespondedAt.Format(time.RFC3339)
		resp.RespondedAt = &respondedAt
	}

	if invitation.TTR != nil {
		ttrResp := convertTTRToResponse(invitation.TTR)
		resp.TTR = &ttrResp
	}

	if invitation.InviterUser != nil {
		userResp := convertUserToResponse(invitation.InviterUser)
		resp.InviterUser = &userResp
	}

	if invitation.InviteeUser != nil {
		userResp := convertUserToResponse(invitation.InviteeUser)
		resp.InviteeUser = &userResp
	}

	return resp
}
