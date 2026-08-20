package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"starter-backend/internal/respond"
	pricingservice "starter-backend/internal/services"

	"github.com/gorilla/mux"
)

type PricingHandler struct {
	pricing *pricingservice.PricingService
}

type createPriceRequest struct {
	ServiceID     uint64 `json:"service_id"`
	VehicleTypeID uint64 `json:"vehicle_type_id"`
	Price         string `json:"price"`
}

type updatePriceRequest struct {
	Price string `json:"price"`
}

func NewPricingHandler(pricing *pricingservice.PricingService) *PricingHandler {
	return &PricingHandler{pricing: pricing}
}

func (h *PricingHandler) Matrix(w http.ResponseWriter, r *http.Request) {
	filter := pricingservice.PricingMatrixFilter{State: strings.TrimSpace(r.URL.Query().Get("state"))}
	var ok bool
	if filter.CategoryID, ok = optionalQueryID(w, r, "category_id"); !ok {
		return
	}
	if filter.ServiceID, ok = optionalQueryID(w, r, "service_id"); !ok {
		return
	}
	if filter.VehicleTypeID, ok = optionalQueryID(w, r, "vehicle_type_id"); !ok {
		return
	}
	switch filter.State {
	case "", "all", "configured", "missing":
	default:
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid price state filter", map[string]string{"state": "Use all, configured or missing"})
		return
	}
	matrix, err := h.pricing.Matrix(r.Context(), filter)
	if err != nil {
		if errors.Is(err, pricingservice.ErrPricingMatrixTooLarge) {
			respond.Error(w, r, http.StatusUnprocessableEntity, "PRICING_MATRIX_TOO_LARGE", "Filter the matrix before loading more than 10000 combinations", nil)
			return
		}
		respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Pricing matrix could not be loaded", nil)
		return
	}
	respond.JSON(w, http.StatusOK, matrix, "Pricing matrix loaded")
}

func (h *PricingHandler) Create(w http.ResponseWriter, r *http.Request) {
	var request createPriceRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	if request.ServiceID == 0 || request.VehicleTypeID == 0 || strings.TrimSpace(request.Price) == "" {
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Service, vehicle type and price are required", map[string]string{
			"service_id": "Select a service", "vehicle_type_id": "Select a vehicle type", "price": "Enter a price",
		})
		return
	}
	item, err := h.pricing.CreatePrice(r.Context(), request.ServiceID, request.VehicleTypeID, request.Price)
	if err != nil {
		handlePricingWriteError(w, r, err)
		return
	}
	respond.JSON(w, http.StatusCreated, priceResponse(item.ServiceID, item.VehicleTypeID, item.Price, item.UpdatedAt), "Service price created")
}

func (h *PricingHandler) Update(w http.ResponseWriter, r *http.Request) {
	serviceID, vehicleTypeID, ok := pricingPathIDs(w, r)
	if !ok {
		return
	}
	var request updatePriceRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	if strings.TrimSpace(request.Price) == "" {
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Price is required", map[string]string{"price": "Enter a price"})
		return
	}
	item, err := h.pricing.UpdatePrice(r.Context(), serviceID, vehicleTypeID, request.Price)
	if err != nil {
		handlePricingWriteError(w, r, err)
		return
	}
	respond.JSON(w, http.StatusOK, priceResponse(item.ServiceID, item.VehicleTypeID, item.Price, item.UpdatedAt), "Service price updated")
}

func (h *PricingHandler) Current(w http.ResponseWriter, r *http.Request) {
	serviceID, ok := requiredQueryID(w, r, "service_id")
	if !ok {
		return
	}
	vehicleTypeID, ok := requiredQueryID(w, r, "vehicle_type_id")
	if !ok {
		return
	}
	item, err := h.pricing.CurrentPrice(r.Context(), serviceID, vehicleTypeID)
	if err != nil {
		if errors.Is(err, pricingservice.ErrPriceNotFound) {
			respond.Error(w, r, http.StatusNotFound, "CURRENT_PRICE_NOT_FOUND", "No current price exists for this active combination", nil)
			return
		}
		respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Current price could not be loaded", nil)
		return
	}
	respond.JSON(w, http.StatusOK, priceResponse(item.ServiceID, item.VehicleTypeID, item.Price, item.UpdatedAt), "Current price loaded")
}

func handlePricingWriteError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, pricingservice.ErrInvalidPrice):
		respond.Error(w, r, http.StatusUnprocessableEntity, "INVALID_PRICE", "Price must be a plain decimal with at most two fractional digits", map[string]string{"price": "Use a value such as 25.00"})
	case errors.Is(err, pricingservice.ErrPriceMustBePositive):
		respond.Error(w, r, http.StatusUnprocessableEntity, "PRICE_MUST_BE_POSITIVE", "Price must be greater than zero", map[string]string{"price": "Enter an amount greater than 0.00"})
	case errors.Is(err, pricingservice.ErrCombinationUnavailable):
		respond.Error(w, r, http.StatusUnprocessableEntity, "PRICING_COMBINATION_UNAVAILABLE", "Service, category and vehicle type must exist and be active", nil)
	case errors.Is(err, pricingservice.ErrPriceAlreadyExists):
		respond.Error(w, r, http.StatusConflict, "SERVICE_PRICE_EXISTS", "A price already exists for this service and vehicle type", nil)
	case errors.Is(err, pricingservice.ErrPriceNotFound):
		respond.Error(w, r, http.StatusNotFound, "SERVICE_PRICE_NOT_FOUND", "Service price was not found", nil)
	default:
		respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Service price could not be saved", nil)
	}
}

func pricingPathIDs(w http.ResponseWriter, r *http.Request) (uint64, uint64, bool) {
	variables := mux.Vars(r)
	serviceID, valid := queryUint(variables["service_id"])
	if !valid {
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid service identifier", nil)
		return 0, 0, false
	}
	vehicleTypeID, valid := queryUint(variables["vehicle_type_id"])
	if !valid {
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid vehicle type identifier", nil)
		return 0, 0, false
	}
	return serviceID, vehicleTypeID, true
}

func requiredQueryID(w http.ResponseWriter, r *http.Request, name string) (uint64, bool) {
	value := strings.TrimSpace(r.URL.Query().Get(name))
	if value == "" {
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Required identifier is missing", map[string]string{name: "Required"})
		return 0, false
	}
	id, valid := queryUint(value)
	if !valid {
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid identifier", map[string]string{name: "Invalid identifier"})
		return 0, false
	}
	return id, true
}

func optionalQueryID(w http.ResponseWriter, r *http.Request, name string) (uint64, bool) {
	value := strings.TrimSpace(r.URL.Query().Get(name))
	if value == "" {
		return 0, true
	}
	id, valid := queryUint(value)
	if !valid {
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid filter identifier", map[string]string{name: "Invalid identifier"})
		return 0, false
	}
	return id, true
}

func priceResponse(serviceID, vehicleTypeID uint64, price string, updatedAt time.Time) map[string]any {
	return map[string]any{
		"service_id": serviceID, "vehicle_type_id": vehicleTypeID, "price": price, "updated_at": updatedAt,
	}
}
