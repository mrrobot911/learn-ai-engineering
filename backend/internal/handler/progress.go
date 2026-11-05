package handler

import (
	"net/http"

	"github.com/mrrobot911/learn-ai-engineering/internal/apperr"
	"github.com/mrrobot911/learn-ai-engineering/internal/service"
	"github.com/mrrobot911/learn-ai-engineering/internal/session"
)

// ProgressHandler handles the lesson progress of the logged-in user.
type ProgressHandler struct {
	service *service.ProgressService
}

// NewProgressHandler creates a progress handler.
func NewProgressHandler(progressService *service.ProgressService) *ProgressHandler {
	return &ProgressHandler{service: progressService}
}

// GetMyProgress writes the completed lesson ids of the logged-in user. It takes
// no user id from the request.
func (h *ProgressHandler) GetMyProgress(w http.ResponseWriter, r *http.Request) {
	auth, ok := session.From(r.Context())
	if !ok {
		WriteError(w, &apperr.Error{Code: apperr.CodeAuthentication, Message: "Authentication is required."})
		return
	}

	ids, err := h.service.ListCompletedLessonIDs(r.Context(), auth.User.ID)
	if err != nil {
		WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ids)
}

// MarkLessonComplete marks the lesson complete for the logged-in user and
// writes the completed lesson ids. It takes no user id from the request.
func (h *ProgressHandler) MarkLessonComplete(w http.ResponseWriter, r *http.Request) {
	auth, ok := session.From(r.Context())
	if !ok {
		WriteError(w, &apperr.Error{Code: apperr.CodeAuthentication, Message: "Authentication is required."})
		return
	}

	ids, err := h.service.MarkLessonComplete(r.Context(), auth.User.ID, r.PathValue("lessonId"))
	if err != nil {
		WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ids)
}

// UnmarkLessonComplete removes the lesson completion for the logged-in user and
// writes the completed lesson ids. It takes no user id from the request.
func (h *ProgressHandler) UnmarkLessonComplete(w http.ResponseWriter, r *http.Request) {
	auth, ok := session.From(r.Context())
	if !ok {
		WriteError(w, &apperr.Error{Code: apperr.CodeAuthentication, Message: "Authentication is required."})
		return
	}

	ids, err := h.service.UnmarkLessonComplete(r.Context(), auth.User.ID, r.PathValue("lessonId"))
	if err != nil {
		WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ids)
}
