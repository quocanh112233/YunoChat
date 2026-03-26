package http

import (
	"encoding/json"
	"net/http"

	"backend/internal/pkg/response"
	"backend/internal/usecase/upload"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type UploadHandler struct {
	uploadUC upload.Usecase
}

func NewUploadHandler(uc upload.Usecase) *UploadHandler {
	return &UploadHandler{uploadUC: uc}
}

func (h *UploadHandler) RegisterRoutes(r chi.Router) {
	r.Route("/upload", func(r chi.Router) {
		r.Post("/avatar/presign", h.PresignAvatar)
		r.Post("/file/presign", h.PresignFile)
	})
}

type presignAvatarRequest struct {
	FileName string `json:"file_name"`
	MimeType string `json:"mime_type"`
	FileSize int64  `json:"file_size"`
}

func (h *UploadHandler) PresignAvatar(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		response.Err(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}
	userID := userIDVal.(uuid.UUID).String()

	var req presignAvatarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format")
		return
	}

	if req.MimeType == "" || req.FileSize <= 0 {
		response.Err(w, http.StatusBadRequest, "VALIDATION_ERROR", "MimeType and FileSize are required")
		return
	}

	res, err := h.uploadUC.PresignAvatar(r.Context(), userID, req.MimeType, req.FileSize)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "UPLOAD_PRESIGN_ERROR", err.Error())
		return
	}

	response.OK(w, http.StatusOK, res)
}

type presignFileRequest struct {
	ConversationID string `json:"conversation_id"`
	FileType       string `json:"file_type"`
	FileName       string `json:"file_name"`
	MimeType       string `json:"mime_type"`
	FileSize       int64  `json:"file_size"`
}

func (h *UploadHandler) PresignFile(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		response.Err(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}
	userID := userIDVal.(uuid.UUID).String()

	var req presignFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format")
		return
	}

	if req.ConversationID == "" || req.FileType == "" || req.MimeType == "" || req.FileSize <= 0 {
		response.Err(w, http.StatusBadRequest, "VALIDATION_ERROR", "Missing required fields")
		return
	}

	res, err := h.uploadUC.PresignFile(r.Context(), userID, req.ConversationID, req.FileType, req.MimeType, req.FileName, req.FileSize)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "UPLOAD_PRESIGN_ERROR", err.Error())
		return
	}

	response.OK(w, http.StatusOK, res)
}
