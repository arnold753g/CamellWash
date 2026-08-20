package handlers

import (
	"errors"
	"net/http"
	"strings"

	"starter-backend/internal/respond"
	customerservice "starter-backend/internal/services"
)

type CustomerHandler struct {
	customers *customerservice.CustomerService
}

type clientRequest struct {
	FullName string  `json:"full_name"`
	CI       *string `json:"ci"`
	Phone    string  `json:"phone"`
}

type vehicleRequest struct {
	VehicleTypeID uint64  `json:"vehicle_type_id"`
	Plate         *string `json:"plate"`
	Brand         string  `json:"brand"`
	Model         string  `json:"model"`
	Color         string  `json:"color"`
}

func NewCustomerHandler(customers *customerservice.CustomerService) *CustomerHandler {
	return &CustomerHandler{customers: customers}
}

func (h *CustomerHandler) ListClients(w http.ResponseWriter, r *http.Request) {
	page, limit := customerPage(r)
	items, pagination, err := h.customers.ListClients(r.Context(), r.URL.Query().Get("search"), page, limit)
	if err != nil {
		handleCustomerError(w, r, err, "Clients could not be loaded")
		return
	}
	respond.JSON(w, http.StatusOK, map[string]any{"clients": items, "pagination": pagination}, "Clients loaded")
}

func (h *CustomerHandler) CreateClient(w http.ResponseWriter, r *http.Request) {
	var request clientRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.customers.CreateClient(r.Context(), customerservice.ClientInput(request))
	if err != nil {
		handleCustomerError(w, r, err, "Client could not be created")
		return
	}
	respond.JSON(w, http.StatusCreated, item, "Client created")
}

func (h *CustomerHandler) GetClient(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	item, err := h.customers.GetClient(r.Context(), id)
	if err != nil {
		handleCustomerError(w, r, err, "Client could not be loaded")
		return
	}
	respond.JSON(w, http.StatusOK, item, "Client loaded")
}

func (h *CustomerHandler) UpdateClient(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var request clientRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.customers.UpdateClient(r.Context(), id, customerservice.ClientInput(request))
	if err != nil {
		handleCustomerError(w, r, err, "Client could not be updated")
		return
	}
	respond.JSON(w, http.StatusOK, item, "Client updated")
}

func (h *CustomerHandler) ListVehicles(w http.ResponseWriter, r *http.Request) {
	clientID, ok := pathID(w, r)
	if !ok {
		return
	}
	page, limit := customerPage(r)
	items, pagination, err := h.customers.ListVehicles(r.Context(), clientID, r.URL.Query().Get("search"), page, limit)
	if err != nil {
		handleCustomerError(w, r, err, "Vehicles could not be loaded")
		return
	}
	respond.JSON(w, http.StatusOK, map[string]any{"vehicles": items, "pagination": pagination}, "Vehicles loaded")
}

func (h *CustomerHandler) CreateVehicle(w http.ResponseWriter, r *http.Request) {
	clientID, ok := pathID(w, r)
	if !ok {
		return
	}
	var request vehicleRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.customers.CreateVehicle(r.Context(), clientID, customerservice.VehicleInput(request))
	if err != nil {
		handleCustomerError(w, r, err, "Vehicle could not be created")
		return
	}
	respond.JSON(w, http.StatusCreated, item, "Vehicle created")
}

func (h *CustomerHandler) GetVehicle(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	item, err := h.customers.GetVehicle(r.Context(), id)
	if err != nil {
		handleCustomerError(w, r, err, "Vehicle could not be loaded")
		return
	}
	respond.JSON(w, http.StatusOK, item, "Vehicle loaded")
}

func (h *CustomerHandler) UpdateVehicle(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var request vehicleRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.customers.UpdateVehicle(r.Context(), id, customerservice.VehicleInput(request))
	if err != nil {
		handleCustomerError(w, r, err, "Vehicle could not be updated")
		return
	}
	respond.JSON(w, http.StatusOK, item, "Vehicle updated")
}

func customerPage(r *http.Request) (int, int) {
	return queryInt(r, "page", 1, 1, 1000000), queryInt(r, "limit", 20, 1, 100)
}

func handleCustomerError(w http.ResponseWriter, r *http.Request, err error, fallback string) {
	var fieldError customerservice.FieldError
	switch {
	case errors.As(err, &fieldError):
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Review the highlighted fields", map[string]string{fieldError.Field: fieldError.Message})
	case errors.Is(err, customerservice.ErrClientNotFound):
		respond.Error(w, r, http.StatusNotFound, "CLIENT_NOT_FOUND", "Client was not found", nil)
	case errors.Is(err, customerservice.ErrVehicleNotFound):
		respond.Error(w, r, http.StatusNotFound, "VEHICLE_NOT_FOUND", "Vehicle was not found", nil)
	case errors.Is(err, customerservice.ErrCIExists):
		respond.Error(w, r, http.StatusConflict, "CLIENT_CI_EXISTS", "A client with that CI already exists", map[string]string{"ci": "CI already exists"})
	case errors.Is(err, customerservice.ErrPlateExists):
		respond.Error(w, r, http.StatusConflict, "VEHICLE_PLATE_EXISTS", "A vehicle with that plate already exists", map[string]string{"plate": "Plate already exists"})
	case errors.Is(err, customerservice.ErrVehicleTypeUnavailable):
		respond.Error(w, r, http.StatusUnprocessableEntity, "VEHICLE_TYPE_UNAVAILABLE", "Vehicle type does not exist or is inactive", map[string]string{"vehicle_type_id": "Select an active vehicle type"})
	default:
		respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", strings.TrimSpace(fallback), nil)
	}
}
