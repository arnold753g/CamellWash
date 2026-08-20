package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"starter-backend/internal/middleware"
	"starter-backend/internal/models"
	"starter-backend/internal/respond"
	orderservice "starter-backend/internal/services"
)

type OrderHandler struct{ orders *orderservice.OrderService }

type createOrderRequest struct {
	VehicleID           uint64    `json:"vehicle_id"`
	ServiceIDs          []uint64  `json:"service_ids"`
	StaffIDs            []uint64  `json:"staff_ids"`
	EstimatedDeliveryAt time.Time `json:"estimated_delivery_at"`
}

type updateOrderRequest struct {
	Version             uint64     `json:"version"`
	VehicleID           *uint64    `json:"vehicle_id"`
	ServiceIDs          *[]uint64  `json:"service_ids"`
	StaffIDs            *[]uint64  `json:"staff_ids"`
	EstimatedDeliveryAt *time.Time `json:"estimated_delivery_at"`
}

type transitionOrderRequest struct {
	Version uint64 `json:"version"`
	Target  string `json:"target"`
	Reason  string `json:"reason"`
}

type cancelOrderRequest struct {
	Version uint64 `json:"version"`
	Reason  string `json:"reason"`
}

func NewOrderHandler(orders *orderservice.OrderService) *OrderHandler {
	return &OrderHandler{orders: orders}
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	actor, ok := orderActor(w, r)
	if !ok {
		return
	}
	var request createOrderRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.orders.Create(r.Context(), actor, orderservice.CreateOrderInput(request))
	if err != nil {
		handleOrderError(w, r, err)
		return
	}
	respond.JSON(w, http.StatusCreated, item, "Order created")
}

func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	actor, ok := orderActor(w, r)
	if !ok {
		return
	}
	filter := orderservice.OrderFilter{Status: strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("status"))),
		Page: queryInt(r, "page", 1, 1, 1000000), Limit: queryInt(r, "limit", 20, 1, 100)}
	if filter.Status != "" && !orderservice.ValidOrderStatus(filter.Status) {
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid order status", map[string]string{"status": "Invalid status"})
		return
	}
	var valid bool
	if filter.VehicleID, valid = optionalQueryID(w, r, "vehicle_id"); !valid {
		return
	}
	if filter.ClientID, valid = optionalQueryID(w, r, "client_id"); !valid {
		return
	}
	if filter.StaffID, valid = optionalQueryID(w, r, "staff_id"); !valid {
		return
	}
	location, err := time.LoadLocation("America/La_Paz")
	if err != nil {
		respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Operational timezone is unavailable", nil)
		return
	}
	if filter.From, valid = orderDateQuery(w, r, "date_from", location, false); !valid {
		return
	}
	if filter.To, valid = orderDateQuery(w, r, "date_to", location, true); !valid {
		return
	}
	if filter.From != nil && filter.To != nil && !filter.From.Before(*filter.To) {
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid date range", nil)
		return
	}
	items, pagination, err := h.orders.List(r.Context(), actor, filter)
	if err != nil {
		handleOrderError(w, r, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]any{"orders": items, "pagination": pagination}, "Orders loaded")
}

func (h *OrderHandler) Get(w http.ResponseWriter, r *http.Request) {
	actor, ok := orderActor(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	item, err := h.orders.Get(r.Context(), actor, id)
	if err != nil {
		handleOrderError(w, r, err)
		return
	}
	respond.JSON(w, http.StatusOK, item, "Order loaded")
}

func (h *OrderHandler) Update(w http.ResponseWriter, r *http.Request) {
	actor, ok := orderActor(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var request updateOrderRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.orders.Update(r.Context(), actor, id, orderservice.UpdateOrderInput(request))
	if err != nil {
		handleOrderError(w, r, err)
		return
	}
	respond.JSON(w, http.StatusOK, item, "Order updated")
}

func (h *OrderHandler) Transition(w http.ResponseWriter, r *http.Request) {
	actor, ok := orderActor(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var request transitionOrderRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.orders.Transition(r.Context(), actor, id, orderservice.TransitionOrderInput(request))
	if err != nil {
		handleOrderError(w, r, err)
		return
	}
	respond.JSON(w, http.StatusOK, item, "Order status updated")
}

func (h *OrderHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	actor, ok := orderActor(w, r)
	if !ok {
		return
	}
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var request cancelOrderRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	item, err := h.orders.Transition(r.Context(), actor, id, orderservice.TransitionOrderInput{
		Version: request.Version, Target: models.OrderCancelled, Reason: request.Reason,
	})
	if err != nil {
		handleOrderError(w, r, err)
		return
	}
	respond.JSON(w, http.StatusOK, item, "Order cancelled")
}

func (h *OrderHandler) AvailableStaff(w http.ResponseWriter, r *http.Request) {
	actor, ok := orderActor(w, r)
	if !ok {
		return
	}
	if actor.Role != models.RoleAdmin && actor.Role != models.RoleManager {
		respond.Error(w, r, http.StatusForbidden, "FORBIDDEN", "You do not have permission for this action", nil)
		return
	}
	items, err := h.orders.AvailableStaff(r.Context())
	if err != nil {
		handleOrderError(w, r, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]any{"staff": items}, "Available wash staff loaded")
}

func (h *OrderHandler) AvailableServices(w http.ResponseWriter, r *http.Request) {
	actor, ok := orderActor(w, r)
	if !ok {
		return
	}
	if actor.Role != models.RoleAdmin && actor.Role != models.RoleManager {
		respond.Error(w, r, http.StatusForbidden, "FORBIDDEN", "You do not have permission for this action", nil)
		return
	}
	vehicleTypeID, ok := requiredQueryID(w, r, "vehicle_type_id")
	if !ok {
		return
	}
	items, err := h.orders.AvailableServices(r.Context(), vehicleTypeID)
	if err != nil {
		handleOrderError(w, r, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]any{"services": items}, "Available priced services loaded")
}

func (h *OrderHandler) ClientHistory(w http.ResponseWriter, r *http.Request) {
	h.listHistory(w, r, true)
}

func (h *OrderHandler) VehicleHistory(w http.ResponseWriter, r *http.Request) {
	h.listHistory(w, r, false)
}

func (h *OrderHandler) listHistory(w http.ResponseWriter, r *http.Request, client bool) {
	actor, ok := orderActor(w, r)
	if !ok {
		return
	}
	if actor.Role != models.RoleAdmin && actor.Role != models.RoleManager {
		respond.Error(w, r, http.StatusForbidden, "FORBIDDEN", "You do not have permission for this action", nil)
		return
	}
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	filter := orderservice.OrderFilter{Page: queryInt(r, "page", 1, 1, 1000000), Limit: queryInt(r, "limit", 20, 1, 100)}
	if client {
		filter.ClientID = id
	} else {
		filter.VehicleID = id
	}
	items, pagination, err := h.orders.List(r.Context(), actor, filter)
	if err != nil {
		handleOrderError(w, r, err)
		return
	}
	respond.JSON(w, http.StatusOK, map[string]any{"orders": items, "pagination": pagination}, "Order history loaded")
}

func orderActor(w http.ResponseWriter, r *http.Request) (models.User, bool) {
	principal, ok := middleware.CurrentPrincipal(r)
	if !ok {
		respond.Error(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication is required", nil)
		return models.User{}, false
	}
	return principal.User, true
}

func orderDateQuery(w http.ResponseWriter, r *http.Request, name string, location *time.Location, exclusiveNextDay bool) (*time.Time, bool) {
	value := strings.TrimSpace(r.URL.Query().Get(name))
	if value == "" {
		return nil, true
	}
	parsed, err := time.ParseInLocation("2006-01-02", value, location)
	if err != nil {
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Date must use YYYY-MM-DD", map[string]string{name: "Invalid date"})
		return nil, false
	}
	if exclusiveNextDay {
		parsed = parsed.AddDate(0, 0, 1)
	}
	return &parsed, true
}

func handleOrderError(w http.ResponseWriter, r *http.Request, err error) {
	var fieldError orderservice.FieldError
	switch {
	case errors.As(err, &fieldError):
		respond.Error(w, r, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Review the highlighted fields", map[string]string{fieldError.Field: fieldError.Message})
	case errors.Is(err, orderservice.ErrOrderNotFound):
		respond.Error(w, r, http.StatusNotFound, "ORDER_NOT_FOUND", "Order was not found", nil)
	case errors.Is(err, orderservice.ErrOrderForbidden):
		respond.Error(w, r, http.StatusForbidden, "FORBIDDEN", "You do not have permission for this order", nil)
	case errors.Is(err, orderservice.ErrOrderConflict):
		respond.Error(w, r, http.StatusConflict, "ORDER_CONFLICT", "Order changed in another operation; reload before retrying", nil)
	case errors.Is(err, orderservice.ErrOrderTerminal):
		respond.Error(w, r, http.StatusConflict, "ORDER_TERMINAL", "Delivered or cancelled orders cannot be reopened", nil)
	case errors.Is(err, orderservice.ErrInvalidTransition):
		respond.Error(w, r, http.StatusUnprocessableEntity, "INVALID_ORDER_TRANSITION", "Requested order transition is not allowed", nil)
	case errors.Is(err, orderservice.ErrOrderPaymentRequired):
		respond.Error(w, r, http.StatusConflict, "PAYMENT_REQUIRED", "A valid full payment is required before delivery", nil)
	case errors.Is(err, orderservice.ErrOrderHasValidPayment):
		respond.Error(w, r, http.StatusConflict, "ORDER_HAS_PAYMENT", "Void the valid payment before cancelling the order", nil)
	case errors.Is(err, orderservice.ErrOrderPaidImmutable):
		respond.Error(w, r, http.StatusConflict, "PAID_ORDER_IMMUTABLE", "Vehicle and services cannot change after payment", nil)
	case errors.Is(err, orderservice.ErrOrderNotEditable):
		respond.Error(w, r, http.StatusConflict, "ORDER_NOT_EDITABLE", "These order fields cannot be changed in the current state", nil)
	case errors.Is(err, orderservice.ErrOrderServiceNoPrice):
		respond.Error(w, r, http.StatusUnprocessableEntity, "ORDER_SERVICE_PRICE_MISSING", "Every selected service must be active and have a current price for the vehicle type", map[string]string{"service_ids": "Missing current price"})
	case errors.Is(err, orderservice.ErrOrderVehicleUnavailable):
		respond.Error(w, r, http.StatusUnprocessableEntity, "ORDER_VEHICLE_UNAVAILABLE", "Vehicle must exist and its type must be active", map[string]string{"vehicle_id": "Vehicle unavailable"})
	case errors.Is(err, orderservice.ErrOrderStaffUnavailable):
		respond.Error(w, r, http.StatusUnprocessableEntity, "ORDER_STAFF_UNAVAILABLE", "All assigned users must be active wash staff", map[string]string{"staff_ids": "Staff unavailable"})
	default:
		respond.Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Order operation could not be completed", nil)
	}
}
