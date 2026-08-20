package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"starter-backend/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrPaymentNotFound            = errors.New("payment not found")
	ErrPaymentForbidden           = errors.New("payment forbidden")
	ErrPaymentAlreadyExists       = errors.New("a valid payment already exists")
	ErrPaymentNotValid            = errors.New("payment is not valid")
	ErrPaymentOrderUnavailable    = errors.New("order cannot be paid")
	ErrPaymentDeliveryState       = errors.New("order must be finished before delivery")
	ErrPaymentDeliveredCorrection = errors.New("delivered payment requires atomic correction")
	ErrPaymentConflict            = errors.New("payment changed concurrently")
	ErrPaymentReviewUnavailable   = errors.New("payment review is unavailable")
	ErrPaymentSelfReview          = errors.New("users cannot review their own payment void")
	ErrPaymentIdempotencyConflict = errors.New("idempotency key was used for another operation")
)

var paymentIdempotencyPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

type PaymentService struct {
	db     *gorm.DB
	orders *OrderService
}

type RegisterPaymentInput struct {
	Version           uint64
	Method            string
	ExternalReference string
	IdempotencyKey    string
	Deliver           bool
}

type VoidPaymentInput struct {
	Version uint64
	Reason  string
}

type CorrectPaymentInput struct {
	Version           uint64
	Method            string
	ExternalReference string
	IdempotencyKey    string
	Reason            string
}

type ReviewPaymentInput struct {
	Decision string
	Note     string
}

type PaymentResult struct {
	Payment models.Payment `json:"payment"`
	Order   models.Order   `json:"order"`
}

func NewPaymentService(db *gorm.DB) *PaymentService {
	return &PaymentService{db: db, orders: NewOrderService(db)}
}

func ValidPaymentMethod(value string) bool {
	switch value {
	case models.PaymentMethodCash, models.PaymentMethodQR, models.PaymentMethodBankTransfer, models.PaymentMethodCard:
		return true
	default:
		return false
	}
}

func normalizePaymentFields(method, externalReference, idempotencyKey string) (string, *string, string, error) {
	method = strings.ToUpper(strings.TrimSpace(method))
	if !ValidPaymentMethod(method) {
		return "", nil, "", FieldError{Field: "method", Message: "Select a valid payment method"}
	}
	reference := strings.Join(strings.Fields(externalReference), " ")
	if reference != "" && (len([]rune(reference)) < 3 || len([]rune(reference)) > 120) {
		return "", nil, "", FieldError{Field: "external_reference", Message: "Reference must contain between 3 and 120 characters"}
	}
	var referencePointer *string
	if reference != "" {
		referencePointer = &reference
	}
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if len(idempotencyKey) < 16 || len(idempotencyKey) > 100 || !paymentIdempotencyPattern.MatchString(idempotencyKey) {
		return "", nil, "", FieldError{Field: "idempotency_key", Message: "A valid idempotency key is required"}
	}
	return method, referencePointer, idempotencyKey, nil
}

func normalizePaymentReason(field, value string) (string, error) {
	value = strings.Join(strings.Fields(value), " ")
	if len([]rune(value)) < 5 || len([]rune(value)) > 500 {
		return "", FieldError{Field: field, Message: "Reason must contain between 5 and 500 characters"}
	}
	return value, nil
}

func paymentActorsOnly(actor models.User) error {
	if actor.Role != models.RoleAdmin && actor.Role != models.RoleManager {
		return ErrPaymentForbidden
	}
	return nil
}

func (s *PaymentService) Register(ctx context.Context, actor models.User, orderID uint64, input RegisterPaymentInput) (PaymentResult, bool, error) {
	if err := paymentActorsOnly(actor); err != nil {
		return PaymentResult{}, false, err
	}
	method, reference, idempotencyKey, err := normalizePaymentFields(input.Method, input.ExternalReference, input.IdempotencyKey)
	if err != nil {
		return PaymentResult{}, false, err
	}
	if input.Version == 0 {
		return PaymentResult{}, false, FieldError{Field: "version", Message: "Order version is required"}
	}
	var payment models.Payment
	created := false
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		order, err := lockPaymentOrder(tx, orderID)
		if err != nil {
			return err
		}
		var existing models.Payment
		existingError := tx.Where("idempotency_key = ?", idempotencyKey).First(&existing).Error
		if existingError == nil {
			if existing.OrderID != orderID || existing.Method != method || !sameOptionalText(existing.ExternalReference, reference) || existing.DeliveryRequested != input.Deliver {
				return ErrPaymentIdempotencyConflict
			}
			payment = existing
			return nil
		}
		if !errors.Is(existingError, gorm.ErrRecordNotFound) {
			return existingError
		}
		if order.Version != input.Version {
			return ErrPaymentConflict
		}
		if order.Status == models.OrderCancelled || order.Status == models.OrderDelivered {
			return ErrPaymentOrderUnavailable
		}
		if input.Deliver && order.Status != models.OrderFinished {
			return ErrPaymentDeliveryState
		}
		hasPayment, err := hasValidPayment(tx, orderID)
		if err != nil {
			return err
		}
		if hasPayment {
			return ErrPaymentAlreadyExists
		}
		payment = models.Payment{OrderID: orderID, Amount: order.Total, Method: method, ExternalReference: reference,
			Status: models.PaymentValid, IdempotencyKey: idempotencyKey, DeliveryRequested: input.Deliver,
			CollectedByUserID: actor.ID, PaidAt: time.Now().UTC(), ReviewStatus: models.PaymentReviewNotRequired}
		if err := tx.Create(&payment).Error; err != nil {
			return fmt.Errorf("register payment: %w", err)
		}
		if err := createPaymentEvent(tx, payment.ID, orderID, "REGISTERED", actor.ID, nil); err != nil {
			return err
		}
		if input.Deliver {
			if err := deliverPaidOrder(tx, order, actor.ID); err != nil {
				return err
			}
		}
		created = true
		return nil
	})
	if err != nil {
		return PaymentResult{}, false, err
	}
	result, err := s.result(ctx, actor, payment.ID, orderID)
	return result, created, err
}

func (s *PaymentService) ListForOrder(ctx context.Context, actor models.User, orderID uint64) ([]models.Payment, error) {
	if err := paymentActorsOnly(actor); err != nil {
		return nil, err
	}
	if _, err := s.orders.Get(ctx, actor, orderID); err != nil {
		return nil, err
	}
	items := make([]models.Payment, 0)
	if err := preloadPaymentUsers(s.db.WithContext(ctx)).Where("order_id = ?", orderID).
		Order("paid_at DESC, id DESC").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list order payments: %w", err)
	}
	return items, nil
}

func (s *PaymentService) Void(ctx context.Context, actor models.User, paymentID uint64, input VoidPaymentInput) (PaymentResult, error) {
	if err := paymentActorsOnly(actor); err != nil {
		return PaymentResult{}, err
	}
	reason, err := normalizePaymentReason("reason", input.Reason)
	if err != nil {
		return PaymentResult{}, err
	}
	if input.Version == 0 {
		return PaymentResult{}, FieldError{Field: "version", Message: "Order version is required"}
	}
	var orderID uint64
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var pointer models.Payment
		if err := tx.Select("id", "order_id").First(&pointer, paymentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrPaymentNotFound
			}
			return err
		}
		orderID = pointer.OrderID
		order, err := lockPaymentOrder(tx, orderID)
		if err != nil {
			return err
		}
		if order.Version != input.Version {
			return ErrPaymentConflict
		}
		if order.Status == models.OrderDelivered {
			return ErrPaymentDeliveredCorrection
		}
		var payment models.Payment
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&payment, paymentID).Error; err != nil {
			return ErrPaymentNotFound
		}
		if payment.Status != models.PaymentValid {
			return ErrPaymentNotValid
		}
		now := time.Now().UTC()
		updates := voidPaymentUpdates(actor, reason, now)
		if err := tx.Model(&payment).Updates(updates).Error; err != nil {
			return err
		}
		if err := createPaymentEvent(tx, payment.ID, orderID, "VOIDED", actor.ID, &reason); err != nil {
			return err
		}
		return bumpOrderVersion(tx, order)
	})
	if err != nil {
		return PaymentResult{}, err
	}
	return s.result(ctx, actor, paymentID, orderID)
}

func (s *PaymentService) Correct(ctx context.Context, actor models.User, paymentID uint64, input CorrectPaymentInput) (PaymentResult, bool, error) {
	if err := paymentActorsOnly(actor); err != nil {
		return PaymentResult{}, false, err
	}
	method, reference, idempotencyKey, err := normalizePaymentFields(input.Method, input.ExternalReference, input.IdempotencyKey)
	if err != nil {
		return PaymentResult{}, false, err
	}
	reason, err := normalizePaymentReason("reason", input.Reason)
	if err != nil {
		return PaymentResult{}, false, err
	}
	if input.Version == 0 {
		return PaymentResult{}, false, FieldError{Field: "version", Message: "Order version is required"}
	}
	var orderID uint64
	var replacement models.Payment
	created := false
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var pointer models.Payment
		if err := tx.Select("id", "order_id").First(&pointer, paymentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrPaymentNotFound
			}
			return err
		}
		orderID = pointer.OrderID
		order, err := lockPaymentOrder(tx, orderID)
		if err != nil {
			return err
		}
		var existing models.Payment
		existingError := tx.Where("idempotency_key = ?", idempotencyKey).First(&existing).Error
		if existingError == nil {
			if existing.OrderID != orderID || existing.ReplacesPaymentID == nil || *existing.ReplacesPaymentID != paymentID || existing.Method != method || !sameOptionalText(existing.ExternalReference, reference) {
				return ErrPaymentIdempotencyConflict
			}
			replacement = existing
			return nil
		}
		if !errors.Is(existingError, gorm.ErrRecordNotFound) {
			return existingError
		}
		if order.Version != input.Version {
			return ErrPaymentConflict
		}
		if order.Status == models.OrderCancelled {
			return ErrPaymentOrderUnavailable
		}
		var current models.Payment
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&current, paymentID).Error; err != nil {
			return ErrPaymentNotFound
		}
		if current.Status != models.PaymentValid {
			return ErrPaymentNotValid
		}
		now := time.Now().UTC()
		if err := tx.Model(&current).Updates(voidPaymentUpdates(actor, reason, now)).Error; err != nil {
			return err
		}
		replacementID := current.ID
		replacement = models.Payment{OrderID: orderID, Amount: order.Total, Method: method, ExternalReference: reference,
			Status: models.PaymentValid, IdempotencyKey: idempotencyKey, CollectedByUserID: actor.ID,
			PaidAt: now, ReviewStatus: models.PaymentReviewNotRequired, ReplacesPaymentID: &replacementID}
		if err := tx.Create(&replacement).Error; err != nil {
			return err
		}
		if err := createPaymentEvent(tx, current.ID, orderID, "VOIDED", actor.ID, &reason); err != nil {
			return err
		}
		if err := createPaymentEvent(tx, replacement.ID, orderID, "CORRECTED", actor.ID, &reason); err != nil {
			return err
		}
		if err := bumpOrderVersion(tx, order); err != nil {
			return err
		}
		created = true
		return nil
	})
	if err != nil {
		return PaymentResult{}, false, err
	}
	result, err := s.result(ctx, actor, replacement.ID, orderID)
	return result, created, err
}

func (s *PaymentService) PendingReviews(ctx context.Context, actor models.User, page, limit int) ([]models.Payment, PageData, error) {
	if actor.Role != models.RoleAdmin {
		return nil, PageData{}, ErrPaymentForbidden
	}
	query := s.db.WithContext(ctx).Model(&models.Payment{}).Where("review_status = ?", models.PaymentReviewPending)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, PageData{}, err
	}
	items := make([]models.Payment, 0)
	if err := preloadPaymentUsers(query).Preload("Order").Preload("Order.Vehicle").Preload("Order.Vehicle.Client").
		Order("voided_at, id").Offset((page - 1) * limit).Limit(limit).Find(&items).Error; err != nil {
		return nil, PageData{}, err
	}
	for index := range items {
		items[index].Order.Vehicle.ClientName = items[index].Order.Vehicle.Client.FullName
	}
	return items, pageData(page, limit, total), nil
}

func (s *PaymentService) Review(ctx context.Context, actor models.User, paymentID uint64, input ReviewPaymentInput) (models.Payment, error) {
	if actor.Role != models.RoleAdmin {
		return models.Payment{}, ErrPaymentForbidden
	}
	decision := strings.ToUpper(strings.TrimSpace(input.Decision))
	if decision != models.PaymentReviewReviewed && decision != models.PaymentReviewObserved {
		return models.Payment{}, FieldError{Field: "decision", Message: "Select reviewed or observed"}
	}
	note := strings.Join(strings.Fields(input.Note), " ")
	if len([]rune(note)) > 500 || (decision == models.PaymentReviewObserved && len([]rune(note)) < 5) {
		return models.Payment{}, FieldError{Field: "note", Message: "Observed reviews require a note between 5 and 500 characters"}
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var payment models.Payment
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&payment, paymentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrPaymentNotFound
			}
			return err
		}
		if payment.ReviewStatus != models.PaymentReviewPending || payment.Status != models.PaymentVoided {
			return ErrPaymentReviewUnavailable
		}
		if payment.VoidedByUserID != nil && *payment.VoidedByUserID == actor.ID {
			return ErrPaymentSelfReview
		}
		now := time.Now().UTC()
		var notePointer *string
		if note != "" {
			notePointer = &note
		}
		if err := tx.Model(&payment).Updates(map[string]any{"review_status": decision, "reviewed_by_user_id": actor.ID,
			"reviewed_at": now, "review_note": notePointer}).Error; err != nil {
			return err
		}
		return createPaymentEvent(tx, payment.ID, payment.OrderID, decision, actor.ID, notePointer)
	})
	if err != nil {
		return models.Payment{}, err
	}
	return s.loadPayment(ctx, paymentID)
}

func (s *PaymentService) result(ctx context.Context, actor models.User, paymentID, orderID uint64) (PaymentResult, error) {
	payment, err := s.loadPayment(ctx, paymentID)
	if err != nil {
		return PaymentResult{}, err
	}
	order, err := s.orders.Get(ctx, actor, orderID)
	if err != nil {
		return PaymentResult{}, err
	}
	return PaymentResult{Payment: payment, Order: order}, nil
}

func (s *PaymentService) loadPayment(ctx context.Context, paymentID uint64) (models.Payment, error) {
	var payment models.Payment
	if err := preloadPaymentUsers(s.db.WithContext(ctx)).First(&payment, paymentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Payment{}, ErrPaymentNotFound
		}
		return models.Payment{}, err
	}
	return payment, nil
}

func preloadPaymentUsers(query *gorm.DB) *gorm.DB {
	return query.Preload("CollectedBy").Preload("VoidedBy").Preload("ReviewedBy")
}

func lockPaymentOrder(tx *gorm.DB, orderID uint64) (models.Order, error) {
	var order models.Order
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&order, orderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Order{}, ErrOrderNotFound
		}
		return models.Order{}, err
	}
	return order, nil
}

func hasValidPayment(tx *gorm.DB, orderID uint64) (bool, error) {
	var count int64
	if err := tx.Model(&models.Payment{}).Where("order_id = ? AND status = ?", orderID, models.PaymentValid).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func deliverPaidOrder(tx *gorm.DB, order models.Order, actorID uint64) error {
	if order.Status != models.OrderFinished {
		return ErrPaymentDeliveryState
	}
	result := tx.Model(&models.Order{}).Where("id = ? AND version = ? AND status = ?", order.ID, order.Version, models.OrderFinished).
		Updates(map[string]any{"status": models.OrderDelivered, "version": order.Version + 1})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return ErrPaymentConflict
	}
	from := models.OrderFinished
	reason := "Pago registrado y vehículo entregado"
	return tx.Create(&models.OrderHistory{OrderID: order.ID, FromStatus: &from, ToStatus: models.OrderDelivered,
		ChangedAt: time.Now().UTC(), ActorUserID: actorID, Reason: &reason}).Error
}

func bumpOrderVersion(tx *gorm.DB, order models.Order) error {
	result := tx.Model(&models.Order{}).Where("id = ? AND version = ?", order.ID, order.Version).
		Update("version", order.Version+1)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return ErrPaymentConflict
	}
	return nil
}

func voidPaymentUpdates(actor models.User, reason string, now time.Time) map[string]any {
	updates := map[string]any{"status": models.PaymentVoided, "voided_by_user_id": actor.ID, "voided_at": now, "void_reason": reason}
	if actor.Role == models.RoleAdmin {
		updates["review_status"] = models.PaymentReviewReviewed
		updates["reviewed_by_user_id"] = actor.ID
		updates["reviewed_at"] = now
	} else {
		updates["review_status"] = models.PaymentReviewPending
	}
	return updates
}

func createPaymentEvent(tx *gorm.DB, paymentID, orderID uint64, eventType string, actorID uint64, reason *string) error {
	event := models.PaymentEvent{PaymentID: paymentID, OrderID: orderID, EventType: eventType,
		ActorUserID: actorID, OccurredAt: time.Now().UTC(), Reason: reason}
	if err := tx.Create(&event).Error; err != nil {
		return fmt.Errorf("create payment event: %w", err)
	}
	return nil
}

func sameOptionalText(left, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}
