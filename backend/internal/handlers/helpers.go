package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"starter-backend/internal/models"
	"starter-backend/internal/respond"

	"github.com/go-playground/validator/v10"
)

const maxJSONBody = 1 << 20

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBody)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		respond.Error(w, r, http.StatusBadRequest, "INVALID_JSON", "Request body must be valid JSON", nil)
		return false
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		respond.Error(w, r, http.StatusBadRequest, "INVALID_JSON", "Request body must contain one JSON object", nil)
		return false
	}
	return true
}

func validationFields(err error) map[string]string {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return nil
	}
	fields := make(map[string]string, len(validationErrors))
	for _, item := range validationErrors {
		name := strings.ToLower(item.Field())
		fields[name] = "Invalid value"
	}
	return fields
}

func normalizeEmail(email string) string { return strings.ToLower(strings.TrimSpace(email)) }

func validRole(role string) bool { return models.ValidRole(role) }

func validStatus(status string) bool { return status == "active" || status == "inactive" }
