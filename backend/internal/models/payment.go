package models

import "time"

const (
	PaymentMethodCash         = "CASH"
	PaymentMethodQR           = "QR"
	PaymentMethodBankTransfer = "BANK_TRANSFER"
	PaymentMethodCard         = "CARD"

	PaymentValid  = "VALID"
	PaymentVoided = "VOIDED"

	PaymentReviewNotRequired = "NOT_REQUIRED"
	PaymentReviewPending     = "PENDING_REVIEW"
	PaymentReviewReviewed    = "REVIEWED"
	PaymentReviewObserved    = "OBSERVED"
)

type Payment struct {
	ID                uint64     `gorm:"primaryKey" json:"id"`
	OrderID           uint64     `json:"order_id"`
	Order             Order      `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Amount            string     `gorm:"type:numeric(12,2)" json:"amount"`
	Method            string     `json:"method"`
	ExternalReference *string    `json:"external_reference"`
	Status            string     `json:"status"`
	IdempotencyKey    string     `json:"-"`
	DeliveryRequested bool       `json:"delivery_requested"`
	CollectedByUserID uint64     `json:"collected_by_user_id"`
	CollectedBy       User       `gorm:"foreignKey:CollectedByUserID" json:"collected_by"`
	PaidAt            time.Time  `json:"paid_at"`
	VoidedByUserID    *uint64    `json:"voided_by_user_id"`
	VoidedBy          *User      `gorm:"foreignKey:VoidedByUserID" json:"voided_by,omitempty"`
	VoidedAt          *time.Time `json:"voided_at"`
	VoidReason        *string    `json:"void_reason"`
	ReviewStatus      string     `json:"review_status"`
	ReviewedByUserID  *uint64    `json:"reviewed_by_user_id"`
	ReviewedBy        *User      `gorm:"foreignKey:ReviewedByUserID" json:"reviewed_by,omitempty"`
	ReviewedAt        *time.Time `json:"reviewed_at"`
	ReviewNote        *string    `json:"review_note"`
	ReplacesPaymentID *uint64    `json:"replaces_payment_id"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (Payment) TableName() string { return "payments" }

type PaymentEvent struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	PaymentID   uint64    `json:"payment_id"`
	OrderID     uint64    `json:"order_id"`
	EventType   string    `json:"event_type"`
	ActorUserID uint64    `json:"actor_user_id"`
	Actor       User      `gorm:"foreignKey:ActorUserID" json:"actor"`
	OccurredAt  time.Time `json:"occurred_at"`
	Reason      *string   `json:"reason"`
}

func (PaymentEvent) TableName() string { return "payment_events" }
