package models

import (
	"encoding/json"
	"time"
)

type Document struct {
	ID             uint64          `json:"id" gorm:"primaryKey"`
	DocumentKey    string          `json:"-"`
	Kind           string          `json:"kind"`
	OrderID        uint64          `json:"order_id"`
	PaymentID      *uint64         `json:"payment_id"`
	Snapshot       json.RawMessage `json:"snapshot" gorm:"type:jsonb"`
	IssuedAt       time.Time       `json:"issued_at"`
	IssuedByUserID uint64          `json:"issued_by_user_id"`
}

type DocumentPrintAttempt struct {
	ID          uint64    `json:"id" gorm:"primaryKey"`
	DocumentID  uint64    `json:"document_id"`
	RequestKey  string    `json:"-"`
	CopyNumber  int       `json:"copy_number"`
	ActorUserID uint64    `json:"actor_user_id"`
	RequestedAt time.Time `json:"requested_at"`
}
