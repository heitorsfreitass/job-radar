package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/heitorsfreitass/job-radar/internal/application"
	"github.com/heitorsfreitass/job-radar/internal/domain"
)

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string `json:"token"`
	Email string `json:"email"`
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := application.Register(r.Context(), h.users, req.Email, req.Password)
	switch {
	case errors.Is(err, application.ErrWeakPassword):
		writeError(w, http.StatusBadRequest, err.Error())
		return
	case errors.Is(err, domain.ErrEmailTaken):
		writeError(w, http.StatusConflict, err.Error())
		return
	case err != nil:
		writeError(w, http.StatusInternalServerError, "failed to register")
		return
	}

	h.respondWithToken(w, user)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := application.Login(r.Context(), h.users, req.Email, req.Password)
	switch {
	case errors.Is(err, application.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	case err != nil:
		writeError(w, http.StatusInternalServerError, "failed to log in")
		return
	}

	h.respondWithToken(w, user)
}

func (h *Handler) respondWithToken(w http.ResponseWriter, user *domain.User) {
	token, err := issueToken(h.jwtSecret, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to issue token")
		return
	}
	writeJSON(w, http.StatusOK, authResponse{Token: token, Email: user.Email})
}

func (h *Handler) handleMe(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())

	user, err := h.users.GetByID(r.Context(), userID)
	if err != nil || user == nil {
		writeError(w, http.StatusUnauthorized, "user not found")
		return
	}

	prefs, err := application.GetPreferences(r.Context(), h.users, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load preferences")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"email":       user.Email,
		"preferences": prefs,
	})
}

func (h *Handler) handleSavePreferences(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r.Context())

	var prefs domain.Preferences
	if err := json.NewDecoder(r.Body).Decode(&prefs); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := application.SavePreferences(r.Context(), h.users, userID, prefs); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save preferences")
		return
	}

	writeJSON(w, http.StatusOK, prefs)
}
