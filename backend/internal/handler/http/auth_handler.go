package http

import (
	"encoding/json"
	"net/http"
	"time"

	"backend/internal/pkg/response"
	authuc "backend/internal/usecase/auth"

	"github.com/google/uuid"
)

type AuthHandler struct {
	registerUC *authuc.RegisterUseCase
	loginUC    *authuc.LoginUseCase
	refreshUC  *authuc.RefreshTokenUseCase
	logoutUC   *authuc.LogoutUseCase
}

func NewAuthHandler(
	reg *authuc.RegisterUseCase,
	log *authuc.LoginUseCase,
	ref *authuc.RefreshTokenUseCase,
	out *authuc.LogoutUseCase,
) *AuthHandler {
	return &AuthHandler{
		registerUC: reg,
		loginUC:    log,
		refreshUC:  ref,
		logoutUC:   out,
	}
}

func (h *AuthHandler) setCookie(w http.ResponseWriter, token string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Expires:  expires,
		HttpOnly: true,
		Secure:   true, // change to false if testing locally without HTTPS, but spec says true
		SameSite: http.SameSiteStrictMode,
		Path:     "/v1/auth",
	})
}

func (h *AuthHandler) clearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/v1/auth",
	})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input authuc.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON body")
		return
	}

	out, err, valErrs := h.registerUC.Execute(r.Context(), input)
	if valErrs != nil {
		response.Err(w, http.StatusBadRequest, "VALIDATION_FAILED", "Invalid input data", valErrs)
		return
	}
	if err != nil {
		response.Err(w, http.StatusConflict, "REGISTER_FAILED", err.Error())
		return
	}

	h.setCookie(w, out.RefreshToken, time.Now().Add(7*24*time.Hour))
	response.OK(w, http.StatusCreated, map[string]interface{}{
		"user":         out.User,
		"access_token": out.AccessToken,
		"expires_in":   out.ExpiresIn,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input authuc.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Err(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON format")
		return
	}

	out, err, valErrs := h.loginUC.Execute(r.Context(), input)
	if valErrs != nil {
		response.Err(w, http.StatusBadRequest, "VALIDATION_FAILED", "Invalid input data", valErrs)
		return
	}
	if err != nil {
		// Dùng generic message theo docs số 7 phần 10
		response.Err(w, http.StatusUnauthorized, "LOGIN_FAILED", "Email hoặc mật khẩu không đúng")
		return
	}

	h.setCookie(w, out.RefreshToken, time.Now().Add(7*24*time.Hour))
	response.OK(w, http.StatusOK, map[string]interface{}{
		"user":         out.User,
		"access_token": out.AccessToken,
		"expires_in":   out.ExpiresIn,
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		response.Err(w, http.StatusUnauthorized, "MISSING_COOKIE", "No refresh token provided")
		return
	}

	out, err := h.refreshUC.Execute(r.Context(), cookie.Value)
	if err != nil {
		h.clearCookie(w)
		response.Err(w, http.StatusUnauthorized, "REFRESH_FAILED", err.Error())
		return
	}

	h.setCookie(w, out.RefreshToken, time.Now().Add(7*24*time.Hour))
	response.OK(w, http.StatusOK, map[string]interface{}{
		"access_token": out.AccessToken,
		"expires_in":   out.ExpiresIn,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		response.Err(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	userID := userIDVal.(uuid.UUID)

	var refToken string
	cookie, err := r.Cookie("refresh_token")
	if err == nil {
		refToken = cookie.Value
	}

	h.clearCookie(w)
	_ = h.logoutUC.Execute(r.Context(), authuc.LogoutInput{
		UserID:       userID,
		RefreshToken: refToken,
	})

	response.OK(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}
