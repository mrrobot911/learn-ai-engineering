package handler

import (
	"encoding/json"
	"net/http"

	"github.com/mrrobot911/learn-ai-engineering/internal/apperr"
	"github.com/mrrobot911/learn-ai-engineering/internal/service"
	"github.com/mrrobot911/learn-ai-engineering/internal/session"
)

// ProgressHandler handles lesson progress endpoints.
type ProgressHandler struct {
	service *service.ProgressService
}

// NewProgressHandler creates a progress handler.
func NewProgressHandler(progressService *service.ProgressService) *ProgressHandler {
	return &ProgressHandler{
		service: progressService,
	}
}

// MarkLessonComplete marks a lesson as complete for the logged-in user.
func (h *ProgressHandler) MarkLessonComplete(w http.ResponseWriter, r *http.Request) {
	// Extract lesson ID from path
	lessonID := r.PathValue("lessonId")
	if lessonID == "" {
		WriteError(w, &apperr.Error{
			Code:    apperr.CodeValidation,
			Message: "lesson ID is required",
		})
		return
	}

	// Verify session
	auth, ok := session.From(r.Context())
	if !ok {
		WriteError(w, &apperr.Error{
			Code:    apperr.CodeAuthentication,
			Message: "no valid session",
		})
		return
	}

	// Mark lesson as complete
	progress, err := h.service.MarkLessonComplete(r.Context(), auth.User.ID, lessonID)
	if err != nil {
		WriteError(w, err)
		return
	}

	// Return progress
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(progress)
}

// UnmarkLessonComplete removes completion for a lesson.
func (h *ProgressHandler) UnmarkLessonComplete(w http.ResponseWriter, r *http.Request) {
	// Extract lesson ID from path
	lessonID := r.PathValue("lessonId")
	if lessonID == "" {
		WriteError(w, &apperr.Error{
			Code:    apperr.CodeValidation,
			Message: "lesson ID is required",
		})
		return
	}

	// Verify session
	auth, ok := session.From(r.Context())
	if !ok {
		WriteError(w, &apperr.Error{
			Code:    apperr.CodeAuthentication,
			Message: "no valid session",
		})
		return
	}

	// Remove completion
	progress, err := h.service.UnmarkLessonComplete(r.Context(), auth.User.ID, lessonID)
	if err != nil {
		WriteError(w, err)
		return
	}

	// Return progress
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(progress)
}

// GetMyProgress returns the lesson progress for the logged-in user.
func (h *ProgressHandler) GetMyProgress(w http.ResponseWriter, r *http.Request) {
	// Verify session
	auth, ok := session.From(r.Context())
	if !ok {
		WriteError(w, &apperr.Error{
			Code:    apperr.CodeAuthentication,
			Message: "no valid session",
		})
		return
	}

	// Get progress
	progress, err := h.service.GetUserProgress(r.Context(), auth.User.ID)
	if err != nil {
		WriteError(w, err)
		return
	}

	// Return progress
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(progress)
}
