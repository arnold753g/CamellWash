package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"starter-backend/internal/config"
	"starter-backend/internal/middleware"
	"starter-backend/internal/models"
	"starter-backend/internal/respond"
	"starter-backend/internal/security"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type AdminHandler struct {
	db         *gorm.DB
	validate   *validator.Validate
	bcryptCost int
}

func NewAdminHandler(db *gorm.DB, cfg config.Config) *AdminHandler {
	return &AdminHandler{db: db, validate: validator.New(), bcryptCost: cfg.PasswordBcryptCost}
}

type createUserRequest struct {
	Email     string `json:"email" validate:"required,email,max=320"`
	Password  string `json:"password" validate:"required"`
	FirstName string `json:"first_name" validate:"required,min=2,max=100"`
	LastName  string `json:"last_name" validate:"required,min=2,max=100"`
	Role      string `json:"role" validate:"required"`
	Status    string `json:"status" validate:"required"`
}

type updateUserRequest struct {
	Email     string `json:"email" validate:"required,email,max=320"`
	FirstName string `json:"first_name" validate:"required,min=2,max=100"`
	LastName  string `json:"last_name" validate:"required,min=2,max=100"`
}

func (h *AdminHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var request createUserRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	request.Email = normalizeEmail(request.Email)
	request.FirstName = strings.TrimSpace(request.FirstName)
	request.LastName = strings.TrimSpace(request.LastName)
	request.Role = strings.TrimSpace(request.Role)
	request.Status = strings.TrimSpace(request.Status)
	if err := h.validate.Struct(request); err != nil {
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Review the highlighted fields", validationFields(err))
		return
	}
	if !validRole(request.Role) {
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Role is invalid", map[string]string{"role": "Invalid role"})
		return
	}
	if !validStatus(request.Status) {
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Status is invalid", map[string]string{"status": "Invalid status"})
		return
	}
	hash, err := security.HashPassword(request.Password, h.bcryptCost)
	if err != nil {
		respond.Error(w, r, http.StatusUnprocessableEntity, "PASSWORD_POLICY", security.ErrPasswordPolicy.Error(), map[string]string{"password": security.ErrPasswordPolicy.Error()})
		return
	}
	user := models.User{Email: request.Email, PasswordHash: hash, FirstName: request.FirstName,
		LastName: request.LastName, Role: request.Role, Status: request.Status}
	if err := h.db.Create(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			respond.Error(w, r, http.StatusConflict, "EMAIL_EXISTS", "An account with that email already exists", map[string]string{"email": "Email is already in use"})
			return
		}
		respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "User could not be created", nil)
		return
	}
	respond.JSON(w, http.StatusCreated, user.Public(), "User created")
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page := queryInt(r, "page", 1, 1, 1000000)
	limit := queryInt(r, "limit", 20, 1, 100)
	search := strings.TrimSpace(r.URL.Query().Get("search"))
	query := h.db.Model(&models.User{})
	if search != "" {
		like := "%" + escapeLike(strings.ToLower(search)) + "%"
		query = query.Where("LOWER(email) LIKE ? ESCAPE '!' OR LOWER(first_name || ' ' || last_name) LIKE ? ESCAPE '!'", like, like)
	}
	if role := strings.TrimSpace(r.URL.Query().Get("role")); role != "" {
		if !validRole(role) {
			respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid role filter", nil)
			return
		}
		query = query.Where("role = ?", role)
	}
	if status := strings.TrimSpace(r.URL.Query().Get("status")); status != "" {
		if !validStatus(status) {
			respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid status filter", nil)
			return
		}
		query = query.Where("status = ?", status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Users could not be loaded", nil)
		return
	}
	var users []models.User
	if err := query.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&users).Error; err != nil {
		respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Users could not be loaded", nil)
		return
	}
	public := make([]models.PublicUser, len(users))
	for index, user := range users {
		public[index] = user.Public()
	}
	respond.JSON(w, http.StatusOK, map[string]any{"users": public, "pagination": map[string]any{
		"page": page, "limit": limit, "total": total, "total_pages": (total + int64(limit) - 1) / int64(limit),
	}}, "Users loaded")
}

func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var user models.User
	if err := h.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respond.Error(w, r, http.StatusNotFound, "NOT_FOUND", "User was not found", nil)
			return
		}
		respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "User could not be loaded", nil)
		return
	}
	respond.JSON(w, http.StatusOK, user.Public(), "User loaded")
}

func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var request updateUserRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	request.Email = normalizeEmail(request.Email)
	request.FirstName = strings.TrimSpace(request.FirstName)
	request.LastName = strings.TrimSpace(request.LastName)
	if err := h.validate.Struct(request); err != nil {
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Review the highlighted fields", validationFields(err))
		return
	}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		var current models.User
		if err := tx.First(&current, id).Error; err != nil {
			return err
		}
		if err := tx.Model(&current).Updates(map[string]any{
			"email": request.Email, "first_name": request.FirstName, "last_name": request.LastName,
		}).Error; err != nil {
			return err
		}
		if current.Email != request.Email {
			now := time.Now().UTC()
			return tx.Model(&models.Session{}).Where("user_id = ? AND revoked_at IS NULL", id).Update("revoked_at", now).Error
		}
		return nil
	}); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			respond.Error(w, r, http.StatusConflict, "EMAIL_EXISTS", "An account with that email already exists", map[string]string{"email": "Email is already in use"})
			return
		}
		handleAdminMutationError(w, r, err)
		return
	}
	var updated models.User
	if err := h.db.First(&updated, id).Error; err != nil {
		handleAdminMutationError(w, r, err)
		return
	}
	respond.JSON(w, http.StatusOK, updated.Public(), "User updated")
}

func (h *AdminHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok || !h.canManageTarget(w, r, id) {
		return
	}
	var request struct {
		Role string `json:"role"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}
	if !validRole(request.Role) {
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Role is invalid", map[string]string{"role": "Invalid role"})
		return
	}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&models.User{}).Where("id = ?", id).Update("role", request.Role)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		now := time.Now().UTC()
		return tx.Model(&models.Session{}).Where("user_id = ? AND revoked_at IS NULL", id).Update("revoked_at", now).Error
	}); err != nil {
		handleAdminMutationError(w, r, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]any{"id": id, "role": request.Role}, "Role updated; existing sessions revoked")
}

func (h *AdminHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok || !h.canManageTarget(w, r, id) {
		return
	}
	var request struct {
		Status string `json:"status"`
	}
	if !decodeJSON(w, r, &request) {
		return
	}
	if !validStatus(request.Status) {
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Status must be active or inactive", map[string]string{"status": "Invalid status"})
		return
	}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&models.User{}).Where("id = ?", id).Update("status", request.Status)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		if request.Status == models.StatusInactive {
			now := time.Now().UTC()
			return tx.Model(&models.Session{}).Where("user_id = ? AND revoked_at IS NULL", id).Update("revoked_at", now).Error
		}
		return nil
	}); err != nil {
		handleAdminMutationError(w, r, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]any{"id": id, "status": request.Status}, "Status updated")
}

func (h *AdminHandler) canManageTarget(w http.ResponseWriter, r *http.Request, id uint64) bool {
	principal, ok := middleware.CurrentPrincipal(r)
	if !ok {
		respond.Error(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication is required", nil)
		return false
	}
	if principal.User.ID == id {
		respond.Error(w, r, http.StatusConflict, "SELF_ADMIN_CHANGE", "Use profile settings for your own account", nil)
		return false
	}
	return true
}

func handleAdminMutationError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		respond.Error(w, r, http.StatusNotFound, "NOT_FOUND", "User was not found", nil)
		return
	}
	respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "User could not be updated", nil)
}

func pathID(w http.ResponseWriter, r *http.Request) (uint64, bool) {
	id, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil || id == 0 {
		respond.Error(w, r, http.StatusBadRequest, "INVALID_ID", "Resource ID is invalid", nil)
		return 0, false
	}
	return id, true
}

func queryInt(r *http.Request, key string, fallback, min, max int) int {
	value, err := strconv.Atoi(r.URL.Query().Get(key))
	if err != nil || value < min {
		return fallback
	}
	if value > max {
		return max
	}
	return value
}

func escapeLike(value string) string {
	replacer := strings.NewReplacer("!", "!!", "%", "!%", "_", "!_")
	return replacer.Replace(value)
}
