package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"starter-backend/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrOrderNotFound           = errors.New("order not found")
	ErrOrderForbidden          = errors.New("order forbidden")
	ErrOrderConflict           = errors.New("order changed concurrently")
	ErrOrderTerminal           = errors.New("order is terminal")
	ErrInvalidTransition       = errors.New("invalid order transition")
	ErrOrderPaymentRequired    = errors.New("delivery requires a valid payment")
	ErrOrderHasValidPayment    = errors.New("order has a valid payment")
	ErrOrderPaidImmutable      = errors.New("paid order vehicle and services are immutable")
	ErrOrderNotEditable        = errors.New("order fields are not editable in current state")
	ErrOrderServiceNoPrice     = errors.New("order service has no current price")
	ErrOrderVehicleUnavailable = errors.New("order vehicle type is unavailable")
	ErrOrderStaffUnavailable   = errors.New("order staff is unavailable")
)

type OrderService struct{ db *gorm.DB }

type CreateOrderInput struct {
	VehicleID           uint64
	ServiceIDs          []uint64
	StaffIDs            []uint64
	EstimatedDeliveryAt time.Time
}

type UpdateOrderInput struct {
	Version             uint64
	VehicleID           *uint64
	ServiceIDs          *[]uint64
	StaffIDs            *[]uint64
	EstimatedDeliveryAt *time.Time
}

type TransitionOrderInput struct {
	Version uint64
	Target  string
	Reason  string
}

type OrderFilter struct {
	Status    string
	VehicleID uint64
	ClientID  uint64
	StaffID   uint64
	From      *time.Time
	To        *time.Time
	Page      int
	Limit     int
}

type OrderServiceOption struct {
	ServiceID    uint64 `json:"service_id"`
	ServiceName  string `json:"service_name"`
	CategoryName string `json:"category_name"`
	Price        string `json:"price"`
}

func NewOrderService(db *gorm.DB) *OrderService { return &OrderService{db: db} }

func ValidOrderStatus(status string) bool {
	switch status {
	case models.OrderPending, models.OrderInProgress, models.OrderQualityControl, models.OrderFinished, models.OrderDelivered, models.OrderCancelled:
		return true
	default:
		return false
	}
}

func NextOrderStatus(status string) string {
	switch status {
	case models.OrderPending:
		return models.OrderInProgress
	case models.OrderInProgress:
		return models.OrderQualityControl
	case models.OrderQualityControl:
		return models.OrderFinished
	case models.OrderFinished:
		return models.OrderDelivered
	default:
		return ""
	}
}

func (s *OrderService) Create(ctx context.Context, actor models.User, input CreateOrderInput) (models.Order, error) {
	if actor.Role != models.RoleAdmin && actor.Role != models.RoleManager {
		return models.Order{}, ErrOrderForbidden
	}
	now := time.Now().UTC()
	if input.EstimatedDeliveryAt.IsZero() || input.EstimatedDeliveryAt.Before(now) {
		return models.Order{}, FieldError{Field: "estimated_delivery_at", Message: "Estimated delivery cannot be earlier than reception"}
	}
	var created models.Order
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		vehicle, err := activeOrderVehicle(tx, input.VehicleID)
		if err != nil {
			return err
		}
		lines, total, err := buildOrderLines(ctx, tx, vehicle.VehicleTypeID, input.ServiceIDs)
		if err != nil {
			return err
		}
		staff, err := validateOrderStaff(tx, input.StaffIDs)
		if err != nil {
			return err
		}
		created = models.Order{VehicleID: vehicle.ID, RegisteredByUserID: actor.ID, ReceivedAt: now,
			EstimatedDeliveryAt: input.EstimatedDeliveryAt.UTC(), Status: models.OrderPending, Total: total, Version: 1}
		if err := tx.Create(&created).Error; err != nil {
			return fmt.Errorf("create order: %w", err)
		}
		for index := range lines {
			lines[index].OrderID = created.ID
		}
		if err := tx.Create(&lines).Error; err != nil {
			return fmt.Errorf("create order services: %w", err)
		}
		assignments := make([]models.OrderStaff, len(staff))
		for index, userID := range staff {
			assignments[index] = models.OrderStaff{OrderID: created.ID, UserID: userID, AssignedAt: now}
		}
		if err := tx.Create(&assignments).Error; err != nil {
			return fmt.Errorf("create order staff: %w", err)
		}
		reason := "Orden creada"
		history := models.OrderHistory{OrderID: created.ID, ToStatus: models.OrderPending, ActorUserID: actor.ID, Reason: &reason, ChangedAt: now}
		if err := tx.Create(&history).Error; err != nil {
			return fmt.Errorf("create initial history: %w", err)
		}
		return nil
	})
	if err != nil {
		return models.Order{}, err
	}
	return s.Get(ctx, actor, created.ID)
}

func activeOrderVehicle(tx *gorm.DB, id uint64) (models.Vehicle, error) {
	if id == 0 {
		return models.Vehicle{}, FieldError{Field: "vehicle_id", Message: "Select a vehicle"}
	}
	var vehicle models.Vehicle
	err := tx.Raw(`SELECT v.* FROM vehicles v JOIN vehicle_types vt ON vt.id = v.vehicle_type_id
		WHERE v.id = ? AND vt.active = TRUE FOR SHARE OF v, vt`, id).Scan(&vehicle).Error
	if err != nil {
		return models.Vehicle{}, fmt.Errorf("load order vehicle: %w", err)
	}
	if vehicle.ID == 0 {
		return models.Vehicle{}, ErrOrderVehicleUnavailable
	}
	return vehicle, nil
}

func buildOrderLines(ctx context.Context, tx *gorm.DB, vehicleTypeID uint64, serviceIDs []uint64) ([]models.OrderService, string, error) {
	ids, err := uniqueOrderIDs("service_ids", serviceIDs, 50)
	if err != nil {
		return nil, "", err
	}
	pricing := NewPricingService(tx)
	lines := make([]models.OrderService, 0, len(ids))
	var totalCents int64
	for _, serviceID := range ids {
		if err := lockActiveCombination(tx, serviceID, vehicleTypeID); err != nil {
			return nil, "", ErrOrderServiceNoPrice
		}
		price, err := pricing.CurrentPrice(ctx, serviceID, vehicleTypeID)
		if err != nil {
			if errors.Is(err, ErrPriceNotFound) {
				return nil, "", ErrOrderServiceNoPrice
			}
			return nil, "", err
		}
		var service models.Service
		if err := tx.Select("id", "name").First(&service, serviceID).Error; err != nil {
			return nil, "", ErrOrderServiceNoPrice
		}
		cents, err := priceToCents(price.Price)
		if err != nil || totalCents > 999999999999-cents {
			return nil, "", FieldError{Field: "service_ids", Message: "Order total exceeds the supported amount"}
		}
		totalCents += cents
		lines = append(lines, models.OrderService{ServiceID: serviceID, ServiceNameSnapshot: service.Name, PriceSnapshot: price.Price})
	}
	return lines, centsToPrice(totalCents), nil
}

func validateOrderStaff(tx *gorm.DB, staffIDs []uint64) ([]uint64, error) {
	ids, err := uniqueOrderIDs("staff_ids", staffIDs, 25)
	if err != nil {
		return nil, err
	}
	var found []uint64
	if err := tx.Model(&models.User{}).Clauses(clause.Locking{Strength: "SHARE"}).
		Where("id IN ? AND role = ? AND status = ?", ids, models.RoleWashStaff, models.StatusActive).
		Order("id").Pluck("id", &found).Error; err != nil {
		return nil, fmt.Errorf("validate order staff: %w", err)
	}
	if len(found) != len(ids) {
		return nil, ErrOrderStaffUnavailable
	}
	return found, nil
}

func uniqueOrderIDs(field string, values []uint64, maximum int) ([]uint64, error) {
	if len(values) == 0 || len(values) > maximum {
		return nil, FieldError{Field: field, Message: fmt.Sprintf("Select between 1 and %d values", maximum)}
	}
	seen := make(map[uint64]struct{}, len(values))
	result := make([]uint64, 0, len(values))
	for _, value := range values {
		if value == 0 {
			return nil, FieldError{Field: field, Message: "Identifiers must be positive"}
		}
		if _, exists := seen[value]; exists {
			return nil, FieldError{Field: field, Message: "Duplicate selections are not allowed"}
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result, nil
}

func priceToCents(price string) (int64, error) {
	parts := strings.Split(price, ".")
	if len(parts) != 2 || len(parts[1]) != 2 {
		return 0, ErrInvalidPrice
	}
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, err
	}
	fraction, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, err
	}
	return whole*100 + fraction, nil
}

func centsToPrice(cents int64) string { return fmt.Sprintf("%d.%02d", cents/100, cents%100) }

func (s *OrderService) Get(ctx context.Context, actor models.User, id uint64) (models.Order, error) {
	query := s.authorizedQuery(s.db.WithContext(ctx).Model(&models.Order{}), actor)
	var order models.Order
	err := preloadOrder(query).Where("orders.id = ?", id).First(&order).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Order{}, ErrOrderNotFound
	}
	if err != nil {
		return models.Order{}, fmt.Errorf("load order: %w", err)
	}
	order.Vehicle.ClientName = order.Vehicle.Client.FullName
	return order, nil
}

func (s *OrderService) List(ctx context.Context, actor models.User, filter OrderFilter) ([]models.Order, PageData, error) {
	query := s.authorizedQuery(s.db.WithContext(ctx).Model(&models.Order{}), actor)
	if filter.Status != "" {
		query = query.Where("orders.status = ?", filter.Status)
	}
	if filter.VehicleID != 0 {
		query = query.Where("orders.vehicle_id = ?", filter.VehicleID)
	}
	if filter.ClientID != 0 {
		query = query.Joins("JOIN vehicles v_filter ON v_filter.id = orders.vehicle_id AND v_filter.client_id = ?", filter.ClientID)
	}
	if filter.StaffID != 0 && actor.Role != models.RoleWashStaff {
		query = query.Joins("JOIN order_staff os_filter ON os_filter.order_id = orders.id AND os_filter.user_id = ?", filter.StaffID)
	}
	if filter.From != nil {
		query = query.Where("orders.received_at >= ?", filter.From.UTC())
	}
	if filter.To != nil {
		query = query.Where("orders.received_at < ?", filter.To.UTC())
	}
	var total int64
	if err := query.Distinct("orders.id").Count(&total).Error; err != nil {
		return nil, PageData{}, fmt.Errorf("count orders: %w", err)
	}
	var items []models.Order
	if err := preloadOrder(query).Distinct("orders.*").Order("orders.received_at DESC, orders.id DESC").
		Offset((filter.Page - 1) * filter.Limit).Limit(filter.Limit).Find(&items).Error; err != nil {
		return nil, PageData{}, fmt.Errorf("list orders: %w", err)
	}
	for index := range items {
		items[index].Vehicle.ClientName = items[index].Vehicle.Client.FullName
	}
	return items, pageData(filter.Page, filter.Limit, total), nil
}

func (s *OrderService) authorizedQuery(query *gorm.DB, actor models.User) *gorm.DB {
	if actor.Role == models.RoleWashStaff {
		return query.Joins("JOIN order_staff os_access ON os_access.order_id = orders.id AND os_access.user_id = ?", actor.ID)
	}
	return query
}

func preloadOrder(query *gorm.DB) *gorm.DB {
	return query.
		Preload("Vehicle").Preload("Vehicle.VehicleType").Preload("Vehicle.Client").
		Preload("RegisteredBy").Preload("Services").Preload("Staff", func(db *gorm.DB) *gorm.DB { return db.Order("users.first_name, users.last_name") }).
		Preload("History", func(db *gorm.DB) *gorm.DB { return db.Order("changed_at, id") }).Preload("History.Actor")
}

func (s *OrderService) Update(ctx context.Context, actor models.User, id uint64, input UpdateOrderInput) (models.Order, error) {
	if actor.Role != models.RoleAdmin && actor.Role != models.RoleManager {
		return models.Order{}, ErrOrderForbidden
	}
	if input.Version == 0 {
		return models.Order{}, FieldError{Field: "version", Message: "Order version is required"}
	}
	if input.VehicleID == nil && input.ServiceIDs == nil && input.StaffIDs == nil && input.EstimatedDeliveryAt == nil {
		return models.Order{}, FieldError{Field: "order", Message: "Provide at least one change"}
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current models.Order
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&current, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrOrderNotFound
			}
			return err
		}
		if current.Version != input.Version {
			return ErrOrderConflict
		}
		if current.Status != models.OrderPending && current.Status != models.OrderInProgress {
			return ErrOrderNotEditable
		}
		if current.Status == models.OrderInProgress && (input.VehicleID != nil || input.ServiceIDs != nil) {
			return ErrOrderNotEditable
		}
		if input.VehicleID != nil || input.ServiceIDs != nil {
			hasPayment, err := hasValidPayment(tx, id)
			if err != nil {
				return err
			}
			if hasPayment {
				return ErrOrderPaidImmutable
			}
		}
		updates := map[string]any{"version": current.Version + 1}
		vehicleID := current.VehicleID
		if input.VehicleID != nil {
			vehicleID = *input.VehicleID
			updates["vehicle_id"] = vehicleID
		}
		if input.EstimatedDeliveryAt != nil {
			if input.EstimatedDeliveryAt.Before(current.ReceivedAt) {
				return FieldError{Field: "estimated_delivery_at", Message: "Estimated delivery cannot be earlier than reception"}
			}
			updates["estimated_delivery_at"] = input.EstimatedDeliveryAt.UTC()
		}
		if input.VehicleID != nil || input.ServiceIDs != nil {
			vehicle, err := activeOrderVehicle(tx, vehicleID)
			if err != nil {
				return err
			}
			serviceIDs := []uint64{}
			if input.ServiceIDs != nil {
				serviceIDs = *input.ServiceIDs
			} else if err := tx.Model(&models.OrderService{}).Where("order_id = ?", id).Pluck("service_id", &serviceIDs).Error; err != nil {
				return err
			}
			lines, total, err := buildOrderLines(ctx, tx, vehicle.VehicleTypeID, serviceIDs)
			if err != nil {
				return err
			}
			if err := tx.Where("order_id = ?", id).Delete(&models.OrderService{}).Error; err != nil {
				return err
			}
			for index := range lines {
				lines[index].OrderID = id
			}
			if err := tx.Create(&lines).Error; err != nil {
				return err
			}
			updates["total"] = total
		}
		if input.StaffIDs != nil {
			staff, err := validateOrderStaff(tx, *input.StaffIDs)
			if err != nil {
				return err
			}
			if err := tx.Where("order_id = ?", id).Delete(&models.OrderStaff{}).Error; err != nil {
				return err
			}
			assignments := make([]models.OrderStaff, len(staff))
			for index, userID := range staff {
				assignments[index] = models.OrderStaff{OrderID: id, UserID: userID, AssignedAt: time.Now().UTC()}
			}
			if err := tx.Create(&assignments).Error; err != nil {
				return err
			}
		}
		result := tx.Model(&models.Order{}).Where("id = ? AND version = ?", id, current.Version).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrOrderConflict
		}
		return nil
	})
	if err != nil {
		return models.Order{}, err
	}
	return s.Get(ctx, actor, id)
}

func (s *OrderService) Transition(ctx context.Context, actor models.User, id uint64, input TransitionOrderInput) (models.Order, error) {
	input.Target = strings.ToUpper(strings.TrimSpace(input.Target))
	input.Reason = strings.Join(strings.Fields(input.Reason), " ")
	if input.Version == 0 {
		return models.Order{}, FieldError{Field: "version", Message: "Order version is required"}
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current models.Order
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&current, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrOrderNotFound
			}
			return err
		}
		if actor.Role == models.RoleWashStaff {
			var assigned int64
			if err := tx.Model(&models.OrderStaff{}).Where("order_id = ? AND user_id = ?", id, actor.ID).Count(&assigned).Error; err != nil {
				return err
			}
			if assigned != 1 {
				return ErrOrderNotFound
			}
		}
		if current.Version != input.Version {
			return ErrOrderConflict
		}
		if current.Status == models.OrderDelivered || current.Status == models.OrderCancelled {
			return ErrOrderTerminal
		}
		if input.Target == models.OrderCancelled {
			if actor.Role != models.RoleAdmin && actor.Role != models.RoleManager {
				return ErrOrderForbidden
			}
			if current.Status != models.OrderPending && current.Status != models.OrderInProgress && current.Status != models.OrderQualityControl {
				return ErrInvalidTransition
			}
			if len([]rune(input.Reason)) < 5 || len([]rune(input.Reason)) > 500 {
				return FieldError{Field: "reason", Message: "Cancellation reason must contain between 5 and 500 characters"}
			}
			hasPayment, err := hasValidPayment(tx, id)
			if err != nil {
				return err
			}
			if hasPayment {
				return ErrOrderHasValidPayment
			}
		} else {
			if input.Target != NextOrderStatus(current.Status) {
				return ErrInvalidTransition
			}
			if input.Target == models.OrderDelivered {
				hasPayment, err := hasValidPayment(tx, id)
				if err != nil {
					return err
				}
				if !hasPayment {
					return ErrOrderPaymentRequired
				}
			}
			if actor.Role == models.RoleWashStaff {
				if input.Target != models.OrderInProgress && input.Target != models.OrderQualityControl {
					return ErrOrderForbidden
				}
				var count int64
				if err := tx.Model(&models.OrderStaff{}).Where("order_id = ? AND user_id = ?", id, actor.ID).Count(&count).Error; err != nil {
					return err
				}
				if count != 1 {
					return ErrOrderForbidden
				}
			} else if actor.Role != models.RoleAdmin && actor.Role != models.RoleManager {
				return ErrOrderForbidden
			}
			if input.Reason != "" && (len([]rune(input.Reason)) < 3 || len([]rune(input.Reason)) > 500) {
				return FieldError{Field: "reason", Message: "Observation must contain between 3 and 500 characters"}
			}
		}
		result := tx.Model(&models.Order{}).Where("id = ? AND version = ? AND status = ?", id, current.Version, current.Status).
			Updates(map[string]any{"status": input.Target, "version": current.Version + 1})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrOrderConflict
		}
		var reason *string
		if input.Reason != "" {
			reason = &input.Reason
		}
		history := models.OrderHistory{OrderID: id, FromStatus: &current.Status, ToStatus: input.Target,
			ChangedAt: time.Now().UTC(), ActorUserID: actor.ID, Reason: reason}
		if err := tx.Create(&history).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return models.Order{}, err
	}
	return s.Get(ctx, actor, id)
}

func (s *OrderService) AvailableStaff(ctx context.Context) ([]models.PublicUser, error) {
	var users []models.User
	if err := s.db.WithContext(ctx).Where("role = ? AND status = ?", models.RoleWashStaff, models.StatusActive).
		Order("LOWER(first_name), LOWER(last_name), id").Limit(500).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("load available staff: %w", err)
	}
	items := make([]models.PublicUser, len(users))
	for index, user := range users {
		items[index] = user.Public()
	}
	return items, nil
}

func (s *OrderService) AvailableServices(ctx context.Context, vehicleTypeID uint64) ([]OrderServiceOption, error) {
	if vehicleTypeID == 0 {
		return nil, FieldError{Field: "vehicle_type_id", Message: "Vehicle type is required"}
	}
	items := make([]OrderServiceOption, 0)
	err := s.db.WithContext(ctx).Table("service_prices sp").
		Select("s.id AS service_id, s.name AS service_name, sc.name AS category_name, sp.price::TEXT AS price").
		Joins("JOIN services s ON s.id = sp.service_id AND s.active = TRUE").
		Joins("JOIN service_categories sc ON sc.id = s.category_id AND sc.active = TRUE").
		Joins("JOIN vehicle_types vt ON vt.id = sp.vehicle_type_id AND vt.active = TRUE").
		Where("sp.vehicle_type_id = ?", vehicleTypeID).
		Order("sc.normalized_name, s.normalized_name").Limit(1000).Scan(&items).Error
	if err != nil {
		return nil, fmt.Errorf("load order service options: %w", err)
	}
	return items, nil
}
