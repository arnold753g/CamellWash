package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"starter-backend/internal/models"
)

var ErrDocumentNotFound = errors.New("document not found")

type DocumentLine struct {
	Name  string `json:"name"`
	Price string `json:"price"`
}

// Only receipt fields are serialized: never store an entire user, client or vehicle.
type DocumentSnapshot struct {
	BusinessName        string         `json:"business_name"`
	BusinessAddress     string         `json:"business_address,omitempty"`
	BusinessPhone       string         `json:"business_phone,omitempty"`
	ClientName          string         `json:"client_name"`
	InvoiceRequested    bool           `json:"invoice_requested"`
	Vehicle             string         `json:"vehicle"`
	Plate               string         `json:"plate"`
	ReceivedAt          time.Time      `json:"received_at"`
	EstimatedDeliveryAt time.Time      `json:"estimated_delivery_at"`
	Services            []DocumentLine `json:"services"`
	Total               string         `json:"total"`
	Currency            string         `json:"currency"`
	Method              string         `json:"method,omitempty"`
	PaidAt              *time.Time     `json:"paid_at,omitempty"`
	CollectedBy         string         `json:"collected_by,omitempty"`
	ExternalReference   *string        `json:"external_reference,omitempty"`
}

type DocumentView struct {
	models.Document
	Number            string                        `json:"number"`
	State             string                        `json:"state"`
	ReplacementNumber string                        `json:"replacement_number,omitempty"`
	ReplacesNumber    string                        `json:"replaces_number,omitempty"`
	Attempts          []models.DocumentPrintAttempt `json:"attempts"`
}

type DocumentService struct{ db *gorm.DB }

func NewDocumentService(db *gorm.DB) *DocumentService { return &DocumentService{db: db} }

func documentNumber(d models.Document) string {
	prefix := "R"
	if d.Kind == "RECEPTION" {
		prefix = "F"
	}
	return fmt.Sprintf("%s-%06d", prefix, d.ID)
}

// The caller holds the order row lock. Issuance is idempotent and snapshots are never updated.
func ensureDocument(tx *gorm.DB, actorID, orderID uint64, payment *models.Payment) (models.Document, error) {
	key, kind := fmt.Sprintf("order:%d", orderID), "RECEPTION"
	var paymentID *uint64
	if payment != nil {
		key, kind, paymentID = fmt.Sprintf("payment:%d", payment.ID), "RECEIPT", &payment.ID
	}
	var document models.Document
	err := tx.Where("document_key = ?", key).First(&document).Error
	if err == nil {
		return document, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return document, err
	}
	var order models.Order
	if err := tx.Preload("Vehicle.Client").Preload("Services").First(&order, orderID).Error; err != nil {
		return document, err
	}
	name := strings.TrimSpace(os.Getenv("RECEIPT_BUSINESS_NAME"))
	if name == "" {
		name = "Camell Wash"
	}
	snapshot := DocumentSnapshot{BusinessName: name, BusinessAddress: strings.TrimSpace(os.Getenv("RECEIPT_BUSINESS_ADDRESS")),
		BusinessPhone: strings.TrimSpace(os.Getenv("RECEIPT_BUSINESS_PHONE")), ClientName: order.Vehicle.Client.FullName,
		Vehicle:    strings.TrimSpace(order.Vehicle.Brand + " " + order.Vehicle.Model + " · " + order.Vehicle.Color),
		ReceivedAt: order.ReceivedAt, EstimatedDeliveryAt: order.EstimatedDeliveryAt, Total: order.Total, Currency: "BOB", Services: []DocumentLine{}, InvoiceRequested: order.InvoiceRequested}
	if order.Vehicle.Plate != nil {
		snapshot.Plate = *order.Vehicle.Plate
	}
	for _, line := range order.Services {
		snapshot.Services = append(snapshot.Services, DocumentLine{Name: line.ServiceNameSnapshot, Price: line.PriceSnapshot})
	}
	if payment != nil {
		var collector models.User
		if err := tx.First(&collector, payment.CollectedByUserID).Error; err != nil {
			return document, err
		}
		snapshot.Total, snapshot.Method, snapshot.PaidAt = payment.Amount, payment.Method, &payment.PaidAt
		snapshot.CollectedBy = strings.TrimSpace(collector.FirstName + " " + collector.LastName)
		snapshot.ExternalReference = payment.ExternalReference
		snapshot.InvoiceRequested = payment.InvoiceRequested
	}
	raw, err := json.Marshal(snapshot)
	if err != nil {
		return document, err
	}
	document = models.Document{DocumentKey: key, Kind: kind, OrderID: orderID, PaymentID: paymentID,
		Snapshot: raw, IssuedAt: time.Now().UTC(), IssuedByUserID: actorID}
	err = tx.Create(&document).Error
	return document, err
}

func (s *DocumentService) Issue(ctx context.Context, actor models.User, kind string, id uint64) (DocumentView, error) {
	if err := paymentActorsOnly(actor); err != nil {
		return DocumentView{}, err
	}
	var d models.Document
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		orderID := id
		var p *models.Payment
		if kind == "RECEIPT" {
			var payment models.Payment
			if err := tx.First(&payment, id).Error; err != nil {
				return ErrPaymentNotFound
			}
			orderID, p = payment.OrderID, &payment
		} else if kind != "RECEPTION" {
			return ErrDocumentNotFound
		}
		if _, err := lockPaymentOrder(tx, orderID); err != nil {
			return err
		}
		// Refresh payment after acquiring the shared order lock used by payment corrections.
		if p != nil {
			if err := tx.First(p, p.ID).Error; err != nil {
				return err
			}
		}
		var err error
		d, err = ensureDocument(tx, actor.ID, orderID, p)
		return err
	})
	if err != nil {
		return DocumentView{}, err
	}
	return s.Get(ctx, actor, d.ID)
}

func (s *DocumentService) Get(ctx context.Context, actor models.User, id uint64) (DocumentView, error) {
	if err := paymentActorsOnly(actor); err != nil {
		return DocumentView{}, err
	}
	return documentView(s.db.WithContext(ctx), id)
}

func documentView(tx *gorm.DB, id uint64) (DocumentView, error) {
	var d models.Document
	if err := tx.First(&d, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return DocumentView{}, ErrDocumentNotFound
		}
		return DocumentView{}, err
	}
	view := DocumentView{Document: d, Number: documentNumber(d), State: "VALID", Attempts: []models.DocumentPrintAttempt{}}
	if d.PaymentID != nil {
		var payment models.Payment
		if err := tx.First(&payment, *d.PaymentID).Error; err != nil {
			return view, err
		}
		view.State = payment.Status
		var replacement models.Document
		if err := tx.Joins("JOIN payments p ON p.id = documents.payment_id").Where("p.replaces_payment_id = ?", payment.ID).First(&replacement).Error; err == nil {
			view.State, view.ReplacementNumber = "REPLACED", documentNumber(replacement)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return view, err
		}
		if payment.ReplacesPaymentID != nil {
			var previous models.Document
			if err := tx.Where("payment_id = ?", *payment.ReplacesPaymentID).First(&previous).Error; err == nil {
				view.ReplacesNumber = documentNumber(previous)
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return view, err
			}
		}
	} else {
		var order models.Order
		if err := tx.First(&order, d.OrderID).Error; err != nil {
			return view, err
		}
		view.State = order.Status
	}
	if err := tx.Where("document_id = ?", id).Order("copy_number").Find(&view.Attempts).Error; err != nil {
		return view, err
	}
	return view, nil
}

type DocumentPrintResult struct {
	Document DocumentView                `json:"document"`
	Attempt  models.DocumentPrintAttempt `json:"attempt"`
}

func (s *DocumentService) RequestPrint(ctx context.Context, actor models.User, id uint64, key string) (DocumentPrintResult, error) {
	if err := paymentActorsOnly(actor); err != nil {
		return DocumentPrintResult{}, err
	}
	if len(key) < 16 || len(key) > 100 || !paymentIdempotencyPattern.MatchString(key) {
		return DocumentPrintResult{}, FieldError{Field: "idempotency_key", Message: "A valid print request key is required"}
	}
	var result DocumentPrintResult
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Lock order before document, consistent with payment registration and corrections.
		var pointer models.Document
		if err := tx.First(&pointer, id).Error; err != nil {
			return ErrDocumentNotFound
		}
		if _, err := lockPaymentOrder(tx, pointer.OrderID); err != nil {
			return err
		}
		var d models.Document
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&d, id).Error; err != nil {
			return err
		}
		err := tx.Where("document_id = ? AND request_key = ?", id, key).First(&result.Attempt).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			var count int64
			if err := tx.Model(&models.DocumentPrintAttempt{}).Where("document_id = ?", id).Count(&count).Error; err != nil {
				return err
			}
			result.Attempt = models.DocumentPrintAttempt{DocumentID: id, RequestKey: key, CopyNumber: int(count), ActorUserID: actor.ID, RequestedAt: time.Now().UTC()}
			if err := tx.Create(&result.Attempt).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		result.Document, err = documentView(tx, id)
		return err
	})
	return result, err
}
