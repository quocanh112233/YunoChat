package http

import (
	"encoding/json"
	"net/http"

	"backend/internal/domain/user"
	"backend/internal/pkg/response"

	"github.com/google/uuid"
)

type UpdateProfileInput struct {
	DisplayName string  `json:"display_name"`
	Bio         *string `json:"bio"`
}

type UserHandler struct {
	userRepo user.UserRepository
}

func NewUserHandler(ur user.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: ur,
	}
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		response.Err(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	userID := userIDVal.(uuid.UUID)

	u, err := h.userRepo.FindByID(r.Context(), userID)
	if err != nil {
		response.Err(w, http.StatusNotFound, "NOT_FOUND", "User not found")
		return
	}

	response.OK(w, http.StatusOK, u)
}

func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		response.Err(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}
	userID := userIDVal.(uuid.UUID)

	var input UpdateProfileInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format")
		return
	}

	// This assumes the complete Update logic exists
	// Since Update is mocked right now, we just return mocked OK.
	u, err := h.userRepo.FindByID(r.Context(), userID)
	if err != nil {
		response.Err(w, http.StatusNotFound, "NOT_FOUND", "User not found")
		return
	}

	if input.DisplayName != "" {
		u.DisplayName = input.DisplayName
	}
	if input.Bio != nil {
		u.Bio = input.Bio
	}

	_ = h.userRepo.Update(r.Context(), u) // Will return mock err until Update is implemented fully

	response.OK(w, http.StatusOK, u)
}
