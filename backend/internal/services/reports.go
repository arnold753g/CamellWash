package services

import (
	"context"
	"fmt"
	"time"

	"starter-backend/internal/models"

	"gorm.io/gorm"
)

const ReportTimezone = "America/La_Paz"

type ReportRange struct {
	FromUTC  time.Time `json:"-"`
	ToUTC    time.Time `json:"-"`
	DateFrom string    `json:"date_from"`
	DateTo   string    `json:"date_to"`
	Timezone string    `json:"timezone"`
}

type OperationalSummary struct {
	OrdersReceived      int64 `gorm:"column:orders_received" json:"orders_received"`
	OrdersDelivered     int64 `gorm:"column:orders_delivered" json:"orders_delivered"`
	OrdersCancelled     int64 `gorm:"column:orders_cancelled" json:"orders_cancelled"`
	ActiveOrders        int64 `gorm:"column:active_orders" json:"active_orders"`
	ReadyForPickup      int64 `gorm:"column:ready_for_pickup" json:"ready_for_pickup"`
	OverdueOrders       int64 `gorm:"column:overdue_orders" json:"overdue_orders"`
	AverageCycleMinutes int64 `gorm:"column:average_cycle_minutes" json:"average_cycle_minutes"`
}

type StatusMetric struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

type DailyOperationalMetric struct {
	Date      string `gorm:"column:date" json:"date"`
	Received  int64  `gorm:"column:received" json:"received"`
	Delivered int64  `gorm:"column:delivered" json:"delivered"`
}

type ServiceMetric struct {
	ServiceName string `gorm:"column:service_name" json:"service_name"`
	Uses        int64  `gorm:"column:uses" json:"uses"`
}

type OperationalReport struct {
	Range       ReportRange              `json:"range"`
	Summary     OperationalSummary       `json:"summary"`
	Status      []StatusMetric           `json:"status"`
	Daily       []DailyOperationalMetric `json:"daily"`
	TopServices []ServiceMetric          `json:"top_services"`
}

type FinancialSummary struct {
	Revenue       string `gorm:"column:revenue" json:"revenue"`
	Payments      int64  `gorm:"column:payments" json:"payments"`
	AverageTicket string `gorm:"column:average_ticket" json:"average_ticket"`
	VoidedAmount  string `gorm:"column:voided_amount" json:"voided_amount"`
	VoidedCount   int64  `gorm:"column:voided_count" json:"voided_count"`
}

type PaymentMethodMetric struct {
	Method string `json:"method"`
	Count  int64  `json:"count"`
	Amount string `json:"amount"`
}

type DailyFinancialMetric struct {
	Date     string `gorm:"column:date" json:"date"`
	Payments int64  `gorm:"column:payments" json:"payments"`
	Amount   string `gorm:"column:amount" json:"amount"`
}

type FinancialReport struct {
	Range   ReportRange            `json:"range"`
	Summary FinancialSummary       `json:"summary"`
	Methods []PaymentMethodMetric  `json:"methods"`
	Daily   []DailyFinancialMetric `json:"daily"`
}

type DashboardReport struct {
	Operational OperationalReport `json:"operational"`
	Financial   FinancialReport   `json:"financial"`
}

type ReportService struct{ db *gorm.DB }

func NewReportService(db *gorm.DB) *ReportService { return &ReportService{db: db} }

func (s *ReportService) Operational(ctx context.Context, reportRange ReportRange) (OperationalReport, error) {
	report := OperationalReport{Range: reportRange, Status: make([]StatusMetric, 0, 4), Daily: []DailyOperationalMetric{}, TopServices: []ServiceMetric{}}
	if err := s.db.WithContext(ctx).Raw(`
		SELECT
			(SELECT COUNT(*) FROM orders WHERE received_at >= ? AND received_at < ?) AS orders_received,
			(SELECT COUNT(*) FROM order_status_history WHERE to_status = 'DELIVERED' AND changed_at >= ? AND changed_at < ?) AS orders_delivered,
			(SELECT COUNT(*) FROM order_status_history WHERE to_status = 'CANCELLED' AND changed_at >= ? AND changed_at < ?) AS orders_cancelled,
			(SELECT COUNT(*) FROM orders WHERE status IN ('PENDING','IN_PROGRESS','QUALITY_CONTROL','FINISHED')) AS active_orders,
			(SELECT COUNT(*) FROM orders WHERE status = 'FINISHED') AS ready_for_pickup,
			(SELECT COUNT(*) FROM orders WHERE status IN ('PENDING','IN_PROGRESS','QUALITY_CONTROL','FINISHED') AND estimated_delivery_at < NOW()) AS overdue_orders,
			COALESCE((SELECT ROUND(AVG(EXTRACT(EPOCH FROM (h.changed_at - o.received_at)) / 60))::BIGINT
				FROM order_status_history h JOIN orders o ON o.id = h.order_id
				WHERE h.to_status = 'DELIVERED' AND h.changed_at >= ? AND h.changed_at < ?), 0) AS average_cycle_minutes`,
		reportRange.FromUTC, reportRange.ToUTC, reportRange.FromUTC, reportRange.ToUTC,
		reportRange.FromUTC, reportRange.ToUTC, reportRange.FromUTC, reportRange.ToUTC).Scan(&report.Summary).Error; err != nil {
		return OperationalReport{}, fmt.Errorf("load operational summary: %w", err)
	}

	type statusRow struct {
		Status string
		Count  int64
	}
	var statusRows []statusRow
	if err := s.db.WithContext(ctx).Raw(`SELECT status, COUNT(*) AS count FROM orders
		WHERE status IN ('PENDING','IN_PROGRESS','QUALITY_CONTROL','FINISHED') GROUP BY status`).Scan(&statusRows).Error; err != nil {
		return OperationalReport{}, fmt.Errorf("load order status metrics: %w", err)
	}
	statusCounts := make(map[string]int64, len(statusRows))
	for _, row := range statusRows {
		statusCounts[row.Status] = row.Count
	}
	for _, status := range []string{models.OrderPending, models.OrderInProgress, models.OrderQualityControl, models.OrderFinished} {
		report.Status = append(report.Status, StatusMetric{Status: status, Count: statusCounts[status]})
	}

	if err := s.db.WithContext(ctx).Raw(`
		WITH days AS (
			SELECT generate_series(?::date, ?::date, INTERVAL '1 day')::date AS day
		), received AS (
			SELECT DATE(received_at AT TIME ZONE 'America/La_Paz') AS day, COUNT(*) AS count
			FROM orders WHERE received_at >= ? AND received_at < ? GROUP BY 1
		), delivered AS (
			SELECT DATE(changed_at AT TIME ZONE 'America/La_Paz') AS day, COUNT(*) AS count
			FROM order_status_history WHERE to_status = 'DELIVERED' AND changed_at >= ? AND changed_at < ? GROUP BY 1
		)
		SELECT TO_CHAR(days.day, 'YYYY-MM-DD') AS date, COALESCE(received.count, 0) AS received,
			COALESCE(delivered.count, 0) AS delivered
		FROM days LEFT JOIN received USING (day) LEFT JOIN delivered USING (day) ORDER BY days.day`,
		reportRange.DateFrom, reportRange.DateTo, reportRange.FromUTC, reportRange.ToUTC,
		reportRange.FromUTC, reportRange.ToUTC).Scan(&report.Daily).Error; err != nil {
		return OperationalReport{}, fmt.Errorf("load daily operational metrics: %w", err)
	}

	if err := s.db.WithContext(ctx).Raw(`SELECT os.service_name_snapshot AS service_name, COUNT(*) AS uses
		FROM order_services os JOIN orders o ON o.id = os.order_id
		WHERE o.received_at >= ? AND o.received_at < ? AND o.status <> 'CANCELLED'
		GROUP BY os.service_name_snapshot ORDER BY uses DESC, service_name ASC LIMIT 5`,
		reportRange.FromUTC, reportRange.ToUTC).Scan(&report.TopServices).Error; err != nil {
		return OperationalReport{}, fmt.Errorf("load top services: %w", err)
	}
	return report, nil
}

func (s *ReportService) Financial(ctx context.Context, reportRange ReportRange) (FinancialReport, error) {
	report := FinancialReport{Range: reportRange, Methods: make([]PaymentMethodMetric, 0, 4), Daily: []DailyFinancialMetric{}}
	if err := s.db.WithContext(ctx).Raw(`SELECT
		COALESCE((SELECT SUM(amount) FROM payments WHERE status = 'VALID' AND paid_at >= ? AND paid_at < ?), 0)::NUMERIC(12,2)::TEXT AS revenue,
		(SELECT COUNT(*) FROM payments WHERE status = 'VALID' AND paid_at >= ? AND paid_at < ?) AS payments,
		COALESCE((SELECT ROUND(AVG(amount), 2) FROM payments WHERE status = 'VALID' AND paid_at >= ? AND paid_at < ?), 0)::NUMERIC(12,2)::TEXT AS average_ticket,
		COALESCE((SELECT SUM(amount) FROM payments WHERE status = 'VOIDED' AND voided_at >= ? AND voided_at < ?), 0)::NUMERIC(12,2)::TEXT AS voided_amount,
		(SELECT COUNT(*) FROM payments WHERE status = 'VOIDED' AND voided_at >= ? AND voided_at < ?) AS voided_count`,
		reportRange.FromUTC, reportRange.ToUTC, reportRange.FromUTC, reportRange.ToUTC,
		reportRange.FromUTC, reportRange.ToUTC, reportRange.FromUTC, reportRange.ToUTC,
		reportRange.FromUTC, reportRange.ToUTC).Scan(&report.Summary).Error; err != nil {
		return FinancialReport{}, fmt.Errorf("load financial summary: %w", err)
	}

	type methodRow struct {
		Method string
		Count  int64
		Amount string
	}
	var methodRows []methodRow
	if err := s.db.WithContext(ctx).Raw(`SELECT method, COUNT(*) AS count, SUM(amount)::NUMERIC(12,2)::TEXT AS amount
		FROM payments WHERE status = 'VALID' AND paid_at >= ? AND paid_at < ? GROUP BY method`,
		reportRange.FromUTC, reportRange.ToUTC).Scan(&methodRows).Error; err != nil {
		return FinancialReport{}, fmt.Errorf("load payment method metrics: %w", err)
	}
	methodMap := make(map[string]methodRow, len(methodRows))
	for _, row := range methodRows {
		methodMap[row.Method] = row
	}
	for _, method := range []string{models.PaymentMethodCash, models.PaymentMethodQR, models.PaymentMethodBankTransfer, models.PaymentMethodCard} {
		row := methodMap[method]
		amount := row.Amount
		if amount == "" {
			amount = "0.00"
		}
		report.Methods = append(report.Methods, PaymentMethodMetric{Method: method, Count: row.Count, Amount: amount})
	}

	if err := s.db.WithContext(ctx).Raw(`
		WITH days AS (
			SELECT generate_series(?::date, ?::date, INTERVAL '1 day')::date AS day
		), paid AS (
			SELECT DATE(paid_at AT TIME ZONE 'America/La_Paz') AS day, COUNT(*) AS payments, SUM(amount) AS amount
			FROM payments WHERE status = 'VALID' AND paid_at >= ? AND paid_at < ? GROUP BY 1
		)
		SELECT TO_CHAR(days.day, 'YYYY-MM-DD') AS date, COALESCE(paid.payments, 0) AS payments,
			COALESCE(paid.amount, 0)::NUMERIC(12,2)::TEXT AS amount
		FROM days LEFT JOIN paid USING (day) ORDER BY days.day`,
		reportRange.DateFrom, reportRange.DateTo, reportRange.FromUTC, reportRange.ToUTC).Scan(&report.Daily).Error; err != nil {
		return FinancialReport{}, fmt.Errorf("load daily financial metrics: %w", err)
	}
	return report, nil
}

func (s *ReportService) Dashboard(ctx context.Context, reportRange ReportRange) (DashboardReport, error) {
	operational, err := s.Operational(ctx, reportRange)
	if err != nil {
		return DashboardReport{}, err
	}
	financial, err := s.Financial(ctx, reportRange)
	if err != nil {
		return DashboardReport{}, err
	}
	return DashboardReport{Operational: operational, Financial: financial}, nil
}
