package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"starter-backend/internal/models"

	"gorm.io/gorm"
)

const MaxPricingMatrixCells = 10000

var (
	ErrInvalidPrice           = errors.New("invalid price")
	ErrPriceMustBePositive    = errors.New("price must be greater than zero")
	ErrCombinationUnavailable = errors.New("service and vehicle type must be active")
	ErrPriceAlreadyExists     = errors.New("price already exists")
	ErrPriceNotFound          = errors.New("price not found")
	ErrPricingMatrixTooLarge  = errors.New("pricing matrix is too large")
)

type PricingService struct {
	db *gorm.DB
}

type PricingMatrixFilter struct {
	CategoryID    uint64
	ServiceID     uint64
	VehicleTypeID uint64
	State         string
}

type PricingMatrixCell struct {
	ServiceID       uint64     `json:"service_id"`
	ServiceName     string     `json:"service_name"`
	CategoryID      uint64     `json:"category_id"`
	CategoryName    string     `json:"category_name"`
	VehicleTypeID   uint64     `json:"vehicle_type_id"`
	VehicleTypeName string     `json:"vehicle_type_name"`
	Price           *string    `json:"price"`
	Configured      bool       `json:"configured"`
	UpdatedAt       *time.Time `json:"updated_at,omitempty"`
}

type PricingMatrixSummary struct {
	Total      int64 `json:"total"`
	Configured int64 `json:"configured"`
	Missing    int64 `json:"missing"`
}

type PricingMatrix struct {
	Cells   []PricingMatrixCell  `json:"cells"`
	Summary PricingMatrixSummary `json:"summary"`
}

func NewPricingService(db *gorm.DB) *PricingService {
	return &PricingService{db: db}
}

// NormalizePrice accepts a plain base-10 amount with at most ten integer and
// two fractional digits. It deliberately avoids float parsing so the exact
// decimal supplied to PostgreSQL is also returned by the API.
func NormalizePrice(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || strings.Count(value, ".") > 1 {
		return "", ErrInvalidPrice
	}
	parts := strings.SplitN(value, ".", 2)
	integer := parts[0]
	if integer == "" || len(integer) > 10 || !onlyDigits(integer) {
		return "", ErrInvalidPrice
	}
	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
		if len(fraction) < 1 || len(fraction) > 2 || !onlyDigits(fraction) {
			return "", ErrInvalidPrice
		}
	}
	integer = strings.TrimLeft(integer, "0")
	if integer == "" {
		integer = "0"
	}
	fraction += strings.Repeat("0", 2-len(fraction))
	if integer == "0" && fraction == "00" {
		return "", ErrPriceMustBePositive
	}
	return integer + "." + fraction, nil
}

func onlyDigits(value string) bool {
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}

func (s *PricingService) CreatePrice(ctx context.Context, serviceID, vehicleTypeID uint64, price string) (models.ServicePrice, error) {
	normalized, err := NormalizePrice(price)
	if err != nil {
		return models.ServicePrice{}, err
	}
	item := models.ServicePrice{ServiceID: serviceID, VehicleTypeID: vehicleTypeID, Price: normalized}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockActiveCombination(tx, serviceID, vehicleTypeID); err != nil {
			return err
		}
		if err := tx.Create(&item).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return ErrPriceAlreadyExists
			}
			return fmt.Errorf("create service price: %w", err)
		}
		return nil
	})
	if err != nil {
		return models.ServicePrice{}, err
	}
	return item, nil
}

func (s *PricingService) UpdatePrice(ctx context.Context, serviceID, vehicleTypeID uint64, price string) (models.ServicePrice, error) {
	normalized, err := NormalizePrice(price)
	if err != nil {
		return models.ServicePrice{}, err
	}
	var item models.ServicePrice
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockActiveCombination(tx, serviceID, vehicleTypeID); err != nil {
			return err
		}
		result := tx.Model(&models.ServicePrice{}).
			Where("service_id = ? AND vehicle_type_id = ?", serviceID, vehicleTypeID).
			Update("price", normalized)
		if result.Error != nil {
			return fmt.Errorf("update service price: %w", result.Error)
		}
		if result.RowsAffected != 1 {
			return ErrPriceNotFound
		}
		return tx.Where("service_id = ? AND vehicle_type_id = ?", serviceID, vehicleTypeID).First(&item).Error
	})
	if err != nil {
		return models.ServicePrice{}, err
	}
	return item, nil
}

func lockActiveCombination(tx *gorm.DB, serviceID, vehicleTypeID uint64) error {
	var lockedServiceID uint64
	serviceQuery := `SELECT s.id
		FROM services s
		JOIN service_categories sc ON sc.id = s.category_id
		WHERE s.id = ? AND s.active = TRUE AND sc.active = TRUE
		FOR SHARE OF s, sc`
	if err := tx.Raw(serviceQuery, serviceID).Scan(&lockedServiceID).Error; err != nil {
		return fmt.Errorf("lock service: %w", err)
	}
	if lockedServiceID == 0 {
		return ErrCombinationUnavailable
	}
	var lockedVehicleTypeID uint64
	if err := tx.Raw(`SELECT id FROM vehicle_types WHERE id = ? AND active = TRUE FOR SHARE`, vehicleTypeID).Scan(&lockedVehicleTypeID).Error; err != nil {
		return fmt.Errorf("lock vehicle type: %w", err)
	}
	if lockedVehicleTypeID == 0 {
		return ErrCombinationUnavailable
	}
	return nil
}

// CurrentPrice is the pricing contract for Module 5. It succeeds only when
// the service, its category and the vehicle type are currently active.
func (s *PricingService) CurrentPrice(ctx context.Context, serviceID, vehicleTypeID uint64) (models.ServicePrice, error) {
	var item models.ServicePrice
	err := s.db.WithContext(ctx).
		Table("service_prices AS sp").
		Select("sp.*").
		Joins("JOIN services s ON s.id = sp.service_id AND s.active = TRUE").
		Joins("JOIN service_categories sc ON sc.id = s.category_id AND sc.active = TRUE").
		Joins("JOIN vehicle_types vt ON vt.id = sp.vehicle_type_id AND vt.active = TRUE").
		Where("sp.service_id = ? AND sp.vehicle_type_id = ?", serviceID, vehicleTypeID).
		Take(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.ServicePrice{}, ErrPriceNotFound
	}
	if err != nil {
		return models.ServicePrice{}, fmt.Errorf("load current price: %w", err)
	}
	return item, nil
}

func (s *PricingService) Matrix(ctx context.Context, filter PricingMatrixFilter) (PricingMatrix, error) {
	baseWhere := `s.active = TRUE AND sc.active = TRUE AND vt.active = TRUE
		AND (? = 0 OR sc.id = ?)
		AND (? = 0 OR s.id = ?)
		AND (? = 0 OR vt.id = ?)`
	baseArgs := []any{filter.CategoryID, filter.CategoryID, filter.ServiceID, filter.ServiceID, filter.VehicleTypeID, filter.VehicleTypeID}

	var summary PricingMatrixSummary
	summaryQuery := `SELECT COUNT(*) AS total,
		COUNT(sp.id) AS configured,
		COUNT(*) - COUNT(sp.id) AS missing
		FROM services s
		JOIN service_categories sc ON sc.id = s.category_id
		CROSS JOIN vehicle_types vt
		LEFT JOIN service_prices sp ON sp.service_id = s.id AND sp.vehicle_type_id = vt.id
		WHERE ` + baseWhere
	if err := s.db.WithContext(ctx).Raw(summaryQuery, baseArgs...).Scan(&summary).Error; err != nil {
		return PricingMatrix{}, fmt.Errorf("summarize pricing matrix: %w", err)
	}
	if summary.Total > MaxPricingMatrixCells {
		return PricingMatrix{}, ErrPricingMatrixTooLarge
	}

	stateClause := ""
	switch filter.State {
	case "", "all":
	case "configured":
		stateClause = " AND sp.id IS NOT NULL"
	case "missing":
		stateClause = " AND sp.id IS NULL"
	default:
		return PricingMatrix{}, fmt.Errorf("invalid matrix state: %w", ErrInvalidPrice)
	}
	query := `SELECT
		s.id AS service_id, s.name AS service_name,
		sc.id AS category_id, sc.name AS category_name,
		vt.id AS vehicle_type_id, vt.name AS vehicle_type_name,
		sp.price::TEXT AS price, (sp.id IS NOT NULL) AS configured, sp.updated_at
		FROM services s
		JOIN service_categories sc ON sc.id = s.category_id
		CROSS JOIN vehicle_types vt
		LEFT JOIN service_prices sp ON sp.service_id = s.id AND sp.vehicle_type_id = vt.id
		WHERE ` + baseWhere + stateClause + `
		ORDER BY sc.normalized_name, s.normalized_name, vt.normalized_name
		LIMIT ?`
	args := append(baseArgs, MaxPricingMatrixCells)
	cells := make([]PricingMatrixCell, 0)
	if err := s.db.WithContext(ctx).Raw(query, args...).Scan(&cells).Error; err != nil {
		return PricingMatrix{}, fmt.Errorf("load pricing matrix: %w", err)
	}
	return PricingMatrix{Cells: cells, Summary: summary}, nil
}
