package handlers

import (
	"net/http"
	"strings"

	"starter-backend/internal/middleware"
	"starter-backend/internal/models"
	"starter-backend/internal/respond"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type ProfileHandler struct {
	db       *gorm.DB
	validate *validator.Validate
}

func NewProfileHandler(db *gorm.DB) *ProfileHandler {
	return &ProfileHandler{db: db, validate: validator.New()}
}

type updateProfileRequest struct {
	FirstName string `json:"first_name" validate:"required,min=2,max=100"`
	LastName  string `json:"last_name" validate:"required,min=2,max=100"`
}

func (h *ProfileHandler) Get(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.CurrentPrincipal(r)
	if !ok {
		respond.Error(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication is required", nil)
		return
	}
	respond.JSON(w, http.StatusOK, principal.User.Public(), "Profile loaded")
}

func (h *ProfileHandler) Update(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.CurrentPrincipal(r)
	if !ok {
		respond.Error(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication is required", nil)
		return
	}
	var request updateProfileRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	request.FirstName = strings.TrimSpace(request.FirstName)
	request.LastName = strings.TrimSpace(request.LastName)
	if err := h.validate.Struct(request); err != nil {
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Review the highlighted fields", validationFields(err))
		return
	}
	if err := h.db.Model(&models.User{}).Where("id = ?", principal.User.ID).
		Updates(map[string]any{"first_name": request.FirstName, "last_name": request.LastName}).Error; err != nil {
		respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Profile could not be updated", nil)
		return
	}
	var updated models.User
	if err := h.db.First(&updated, principal.User.ID).Error; err != nil {
		respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Profile could not be loaded", nil)
		return
	}
	respond.JSON(w, http.StatusOK, updated.Public(), "Profile updated")
}
