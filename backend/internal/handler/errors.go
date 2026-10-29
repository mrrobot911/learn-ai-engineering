package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mrrobot911/learn-ai-engineering/internal/apperr"
)

// errorMapping is the HTTP status and default message for a domain error code.
type errorMapping struct {
	status  int
	message string
}

var errorMappings = map[apperr.Code]errorMapping{
	apperr.CodeValidation:     {http.StatusBadRequest, "The request is invalid."},
	apperr.CodeAuthentication: {http.StatusUnauthorized, "Authentication is required."},
	apperr.CodeForbidden:      {http.StatusForbidden, "The request could not be verified. Refresh the page and try again."},
	apperr.CodeAuthorization:  {http.StatusForbidden, "You are not allowed to perform this action."},
	apperr.CodeNotFound:       {http.StatusNotFound, "The requested resource was not found."},
	apperr.CodeConflict:       {http.StatusConflict, "The request conflicts with the current state."},
	apperr.CodeInternal:       {http.StatusInternalServerError, "An unexpected error occurred."},
}

type errorResponse struct {
	Code    apperr.Code    `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// WriteError writes err as the API error response defined by the contract.
// A domain error keeps its code, message and details. Any other error becomes a
// generic internal_error, so the wrapped text never reaches the client.
func WriteError(w http.ResponseWriter, err error) {
	var domainErr *apperr.Error
	if !errors.As(err, &domainErr) {
		writeJSON(w, http.StatusInternalServerError, errorResponse{
			Code:    apperr.CodeInternal,
			Message: errorMappings[apperr.CodeInternal].message,
		})
		return
	}

	mapping, known := errorMappings[domainErr.Code]
	body := errorResponse{
		Code:    domainErr.Code,
		Message: domainErr.Message,
		Details: domainErr.Details,
	}
	if !known {
		mapping = errorMappings[apperr.CodeInternal]
		body.Code = apperr.CodeInternal
	}
	if body.Message == "" {
		body.Message = mapping.message
	}
	writeJSON(w, mapping.status, body)
}

func writeJSON(w http.ResponseWriter, status int, body errorResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}
