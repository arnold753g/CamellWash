package services

import (
	"context"
	"fmt"
	"time"

	"starter-backend/internal/models"
)

type AccountingOrder struct {
	ID               uint64     `json:"id"`
	ReceivedAt       time.Time  `json:"received_at"`
	Status           string     `json:"status"`
	InvoiceRequested bool       `json:"invoice_requested"`
	ClientID         uint64     `json:"client_id"`
	ClientName       string     `json:"client_name"`
	ClientCI         string     `json:"client_ci"`
	ClientPhone      string     `json:"client_phone"`
	ClientEmail      *string    `json:"client_email"`
	Plate            *string    `json:"plate"`
	Total            string     `json:"total"`
	PaymentID        *uint64    `json:"payment_id"`
	PaymentMethod    *string    `json:"payment_method"`
	PaidAt           *time.Time `json:"paid_at"`
}

type AccountingOrdersReport struct {
	Range      ReportRange       `json:"range"`
	Orders     []AccountingOrder `json:"orders"`
	Pagination PageData          `json:"pagination"`
}

func (s *ReportService) AccountingOrders(ctx context.Context, actor models.User, period ReportRange, invoice *bool, page, limit int) (AccountingOrdersReport, error) {
	if actor.Role != models.RoleAdmin && actor.Role != models.RoleAccountant {
		return AccountingOrdersReport{}, ErrOrderForbidden
	}
	query := s.db.WithContext(ctx).Table("orders o").
		Joins("JOIN vehicles v ON v.id = o.vehicle_id").
		Joins("JOIN clients c ON c.id = v.client_id").
		Joins("LEFT JOIN payments p ON p.order_id = o.id AND p.status = ?", models.PaymentValid).
		Where("o.received_at >= ? AND o.received_at < ?", period.FromUTC, period.ToUTC)
	if invoice != nil {
		query = query.Where("o.invoice_requested = ?", *invoice)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return AccountingOrdersReport{}, fmt.Errorf("count accounting orders: %w", err)
	}
	items := make([]AccountingOrder, 0)
	err := query.Select(`o.id, o.received_at, o.status, o.invoice_requested,
		c.id AS client_id, c.full_name AS client_name, COALESCE(c.ci, '00000000') AS client_ci,
		c.phone AS client_phone, c.email AS client_email, v.plate, o.total,
		p.id AS payment_id, p.method AS payment_method, p.paid_at`).
		Order("o.received_at DESC, o.id DESC").Offset((page - 1) * limit).Limit(limit).Scan(&items).Error
	if err != nil {
		return AccountingOrdersReport{}, fmt.Errorf("list accounting orders: %w", err)
	}
	return AccountingOrdersReport{Range: period, Orders: items, Pagination: pageData(page, limit, total)}, nil
}
