package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"starter-backend/internal/models"
	"starter-backend/internal/respond"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type CatalogHandler struct {
	db       *gorm.DB
	validate *validator.Validate
}

func NewCatalogHandler(db *gorm.DB) *CatalogHandler {
	return &CatalogHandler{db: db, validate: validator.New()}
}

type catalogNameRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

type catalogStatusRequest struct {
	Active *bool `json:"active" validate:"required"`
}

type createServiceRequest struct {
	CategoryID uint64 `json:"category_id" validate:"required"`
	Name       string `json:"name" validate:"required,min=2,max=100"`
}

type updateServiceRequest struct {
	CategoryID uint64 `json:"category_id" validate:"required"`
	Name       string `json:"name" validate:"required,min=2,max=100"`
}

func normalizeCatalogDisplayName(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func (h *CatalogHandler) CreateVehicleType(w http.ResponseWriter, r *http.Request) {
	var request catalogNameRequest
	if !h.decodeName(w, r, &request) {
		return
	}
	if !h.catalogNameAvailable(w, r, "vehicle_types", request.Name, 0, 0) {
		return
	}
	item := models.VehicleType{Name: request.Name, Active: true}
	if err := h.db.Create(&item).Error; err != nil {
		handleCatalogWriteError(w, r, err, "vehicle type")
		return
	}
	respond.JSON(w, http.StatusCreated, item, "Vehicle type created")
}

func (h *CatalogHandler) ListVehicleTypes(w http.ResponseWriter, r *http.Request) {
	query := h.db.Model(&models.VehicleType{})
	query, ok := applyCatalogFilters(w, r, query, true)
	if !ok {
		return
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		catalogReadError(w, r, "Vehicle types")
		return
	}
	page, limit := catalogPage(r)
	var items []models.VehicleType
	if err := query.Order("normalized_name ASC").Offset((page - 1) * limit).Limit(limit).Find(&items).Error; err != nil {
		catalogReadError(w, r, "Vehicle types")
		return
	}
	respond.JSON(w, http.StatusOK, catalogPageData("vehicle_types", items, page, limit, total), "Vehicle types loaded")
}

func (h *CatalogHandler) GetVehicleType(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var item models.VehicleType
	if !h.findCatalog(w, r, &item, id, "Vehicle type") {
		return
	}
	respond.JSON(w, http.StatusOK, item, "Vehicle type loaded")
}

func (h *CatalogHandler) UpdateVehicleType(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var request catalogNameRequest
	if !h.decodeName(w, r, &request) {
		return
	}
	if !h.catalogNameAvailable(w, r, "vehicle_types", request.Name, id, 0) {
		return
	}
	result := h.db.Model(&models.VehicleType{}).Where("id = ?", id).Update("name", request.Name)
	if !handleCatalogUpdateResult(w, r, result, "vehicle type") {
		return
	}
	h.GetVehicleType(w, r)
}

func (h *CatalogHandler) UpdateVehicleTypeStatus(w http.ResponseWriter, r *http.Request) {
	h.updateCatalogStatus(w, r, &models.VehicleType{}, "Vehicle type")
}

func (h *CatalogHandler) ActiveVehicleTypes(w http.ResponseWriter, r *http.Request) {
	var items []models.VehicleType
	if err := h.db.Where("active = TRUE").Order("normalized_name ASC").Limit(500).Find(&items).Error; err != nil {
		catalogReadError(w, r, "Vehicle types")
		return
	}
	respond.JSON(w, http.StatusOK, map[string]any{"vehicle_types": items}, "Active vehicle types loaded")
}

func (h *CatalogHandler) CreateServiceCategory(w http.ResponseWriter, r *http.Request) {
	var request catalogNameRequest
	if !h.decodeName(w, r, &request) {
		return
	}
	if !h.catalogNameAvailable(w, r, "service_categories", request.Name, 0, 0) {
		return
	}
	item := models.ServiceCategory{Name: request.Name, Active: true}
	if err := h.db.Create(&item).Error; err != nil {
		handleCatalogWriteError(w, r, err, "service category")
		return
	}
	respond.JSON(w, http.StatusCreated, item, "Service category created")
}

func (h *CatalogHandler) ListServiceCategories(w http.ResponseWriter, r *http.Request) {
	query := h.db.Model(&models.ServiceCategory{})
	query, ok := applyCatalogFilters(w, r, query, true)
	if !ok {
		return
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		catalogReadError(w, r, "Service categories")
		return
	}
	page, limit := catalogPage(r)
	var items []models.ServiceCategory
	if err := query.Order("normalized_name ASC").Offset((page - 1) * limit).Limit(limit).Find(&items).Error; err != nil {
		catalogReadError(w, r, "Service categories")
		return
	}
	respond.JSON(w, http.StatusOK, catalogPageData("service_categories", items, page, limit, total), "Service categories loaded")
}

func (h *CatalogHandler) GetServiceCategory(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var item models.ServiceCategory
	if !h.findCatalog(w, r, &item, id, "Service category") {
		return
	}
	respond.JSON(w, http.StatusOK, item, "Service category loaded")
}

func (h *CatalogHandler) UpdateServiceCategory(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var request catalogNameRequest
	if !h.decodeName(w, r, &request) {
		return
	}
	if !h.catalogNameAvailable(w, r, "service_categories", request.Name, id, 0) {
		return
	}
	result := h.db.Model(&models.ServiceCategory{}).Where("id = ?", id).Update("name", request.Name)
	if !handleCatalogUpdateResult(w, r, result, "service category") {
		return
	}
	h.GetServiceCategory(w, r)
}

func (h *CatalogHandler) UpdateServiceCategoryStatus(w http.ResponseWriter, r *http.Request) {
	h.updateCatalogStatus(w, r, &models.ServiceCategory{}, "Service category")
}

func (h *CatalogHandler) ActiveServiceCategories(w http.ResponseWriter, r *http.Request) {
	var items []models.ServiceCategory
	if err := h.db.Where("active = TRUE").Order("normalized_name ASC").Limit(500).Find(&items).Error; err != nil {
		catalogReadError(w, r, "Service categories")
		return
	}
	respond.JSON(w, http.StatusOK, map[string]any{"service_categories": items}, "Active service categories loaded")
}

func (h *CatalogHandler) CreateService(w http.ResponseWriter, r *http.Request) {
	var request createServiceRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	request.Name = normalizeCatalogDisplayName(request.Name)
	if err := h.validate.Struct(request); err != nil {
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Review the highlighted fields", validationFields(err))
		return
	}
	if !h.categoryExists(w, r, request.CategoryID, true) {
		return
	}
	if !h.catalogNameAvailable(w, r, "services", request.Name, 0, request.CategoryID) {
		return
	}
	item := models.Service{CategoryID: request.CategoryID, Name: request.Name, Active: true}
	if err := h.db.Create(&item).Error; err != nil {
		handleCatalogWriteError(w, r, err, "service")
		return
	}
	if err := h.db.Preload("Category").First(&item, item.ID).Error; err != nil {
		catalogReadError(w, r, "Service")
		return
	}
	respond.JSON(w, http.StatusCreated, item, "Service created")
}

func (h *CatalogHandler) ListServices(w http.ResponseWriter, r *http.Request) {
	query := h.db.Model(&models.Service{})
	query, ok := applyCatalogFilters(w, r, query, true)
	if !ok {
		return
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("category_id")); raw != "" {
		categoryID, valid := queryUint(raw)
		if !valid {
			respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid category filter", nil)
			return
		}
		query = query.Where("category_id = ?", categoryID)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		catalogReadError(w, r, "Services")
		return
	}
	page, limit := catalogPage(r)
	var items []models.Service
	if err := query.Preload("Category").Order("category_id ASC, normalized_name ASC").Offset((page - 1) * limit).Limit(limit).Find(&items).Error; err != nil {
		catalogReadError(w, r, "Services")
		return
	}
	respond.JSON(w, http.StatusOK, catalogPageData("services", items, page, limit, total), "Services loaded")
}

func (h *CatalogHandler) GetService(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var item models.Service
	if err := h.db.Preload("Category").First(&item, id).Error; err != nil {
		handleCatalogFindError(w, r, err, "Service")
		return
	}
	respond.JSON(w, http.StatusOK, item, "Service loaded")
}

func (h *CatalogHandler) UpdateService(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var request updateServiceRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	request.Name = normalizeCatalogDisplayName(request.Name)
	if err := h.validate.Struct(request); err != nil {
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Review the highlighted fields", validationFields(err))
		return
	}
	var current models.Service
	if err := h.db.Select("id", "category_id").First(&current, id).Error; err != nil {
		handleCatalogFindError(w, r, err, "Service")
		return
	}
	if !h.categoryExists(w, r, request.CategoryID, request.CategoryID != current.CategoryID) {
		return
	}
	if !h.catalogNameAvailable(w, r, "services", request.Name, id, request.CategoryID) {
		return
	}
	result := h.db.Model(&models.Service{}).Where("id = ?", id).Updates(map[string]any{"category_id": request.CategoryID, "name": request.Name})
	if !handleCatalogUpdateResult(w, r, result, "service") {
		return
	}
	h.GetService(w, r)
}

func (h *CatalogHandler) UpdateServiceStatus(w http.ResponseWriter, r *http.Request) {
	h.updateCatalogStatus(w, r, &models.Service{}, "Service")
}

func (h *CatalogHandler) ActiveServices(w http.ResponseWriter, r *http.Request) {
	query := h.db.Model(&models.Service{}).
		Joins("JOIN service_categories ON service_categories.id = services.category_id AND service_categories.active = TRUE").
		Where("services.active = TRUE")
	if raw := strings.TrimSpace(r.URL.Query().Get("category_id")); raw != "" {
		categoryID, valid := queryUint(raw)
		if !valid {
			respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid category filter", nil)
			return
		}
		query = query.Where("services.category_id = ?", categoryID)
	}
	var items []models.Service
	if err := query.Preload("Category").Order("services.category_id ASC, services.normalized_name ASC").Limit(1000).Find(&items).Error; err != nil {
		catalogReadError(w, r, "Services")
		return
	}
	respond.JSON(w, http.StatusOK, map[string]any{"services": items}, "Active services loaded")
}

func (h *CatalogHandler) decodeName(w http.ResponseWriter, r *http.Request, request *catalogNameRequest) bool {
	if !decodeJSON(w, r, request) {
		return false
	}
	request.Name = normalizeCatalogDisplayName(request.Name)
	if err := h.validate.Struct(request); err != nil {
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Review the highlighted fields", validationFields(err))
		return false
	}
	return true
}

func (h *CatalogHandler) updateCatalogStatus(w http.ResponseWriter, r *http.Request, model any, label string) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var request catalogStatusRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	if err := h.validate.Struct(request); err != nil {
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Active status is required", map[string]string{"active": "Required"})
		return
	}
	result := h.db.Model(model).Where("id = ?", id).Update("active", *request.Active)
	if !handleCatalogUpdateResult(w, r, result, strings.ToLower(label)) {
		return
	}
	respond.JSON(w, http.StatusOK, map[string]any{"id": id, "active": *request.Active}, label+" status updated")
}

func (h *CatalogHandler) findCatalog(w http.ResponseWriter, r *http.Request, target any, id uint64, label string) bool {
	if err := h.db.First(target, id).Error; err != nil {
		handleCatalogFindError(w, r, err, label)
		return false
	}
	return true
}

func (h *CatalogHandler) categoryExists(w http.ResponseWriter, r *http.Request, id uint64, requireActive bool) bool {
	var count int64
	query := h.db.Model(&models.ServiceCategory{}).Where("id = ?", id)
	if requireActive {
		query = query.Where("active = TRUE")
	}
	if err := query.Count(&count).Error; err != nil {
		catalogReadError(w, r, "Service category")
		return false
	}
	if count != 1 {
		message := "Category does not exist"
		if requireActive {
			message = "Category does not exist or is inactive"
		}
		respond.Error(w, r, http.StatusUnprocessableEntity, "CATEGORY_NOT_AVAILABLE", message, map[string]string{"category_id": message})
		return false
	}
	return true
}

func (h *CatalogHandler) catalogNameAvailable(w http.ResponseWriter, r *http.Request, table, name string, excludeID, categoryID uint64) bool {
	query := h.db.Table(table).Where("normalized_name = normalize_catalog_name(?)", name)
	if excludeID != 0 {
		query = query.Where("id <> ?", excludeID)
	}
	if categoryID != 0 {
		query = query.Where("category_id = ?", categoryID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		catalogReadError(w, r, "Catalog")
		return false
	}
	if count > 0 {
		respond.Error(w, r, http.StatusConflict, "CATALOG_NAME_EXISTS", "A catalog item with that name already exists in this scope", map[string]string{"name": "Name already exists"})
		return false
	}
	return true
}

func applyCatalogFilters(w http.ResponseWriter, r *http.Request, query *gorm.DB, allowStatus bool) (*gorm.DB, bool) {
	if search := normalizeCatalogDisplayName(r.URL.Query().Get("search")); search != "" {
		query = query.Where("normalized_name LIKE ? ESCAPE '!'", "%"+escapeLike(strings.ToLower(search))+"%")
	}
	if allowStatus {
		switch active := strings.TrimSpace(r.URL.Query().Get("active")); active {
		case "", "all":
		case "true":
			query = query.Where("active = TRUE")
		case "false":
			query = query.Where("active = FALSE")
		default:
			respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid active filter", nil)
			return query, false
		}
	}
	return query, true
}

func catalogPage(r *http.Request) (int, int) {
	return queryInt(r, "page", 1, 1, 1000000), queryInt(r, "limit", 20, 1, 100)
}

func catalogPageData(key string, items any, page, limit int, total int64) map[string]any {
	return map[string]any{key: items, "pagination": map[string]any{
		"page": page, "limit": limit, "total": total, "total_pages": (total + int64(limit) - 1) / int64(limit),
	}}
}

func queryUint(value string) (uint64, bool) {
	result, err := strconv.ParseUint(value, 10, 64)
	if err != nil || result == 0 {
		return 0, false
	}
	return result, true
}

func handleCatalogWriteError(w http.ResponseWriter, r *http.Request, err error, label string) {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		respond.Error(w, r, http.StatusConflict, "CATALOG_NAME_EXISTS", "A "+label+" with that name already exists in this scope", map[string]string{"name": "Name already exists"})
		return
	}
	respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Catalog item could not be saved", nil)
}

func handleCatalogUpdateResult(w http.ResponseWriter, r *http.Request, result *gorm.DB, label string) bool {
	if result.Error != nil {
		handleCatalogWriteError(w, r, result.Error, label)
		return false
	}
	if result.RowsAffected != 1 {
		respond.Error(w, r, http.StatusNotFound, "NOT_FOUND", "Catalog item was not found", nil)
		return false
	}
	return true
}

func handleCatalogFindError(w http.ResponseWriter, r *http.Request, err error, label string) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		respond.Error(w, r, http.StatusNotFound, "NOT_FOUND", label+" was not found", nil)
		return
	}
	catalogReadError(w, r, label)
}

func catalogReadError(w http.ResponseWriter, r *http.Request, label string) {
	respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", label+" could not be loaded", nil)
}
