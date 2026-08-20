package models

import "time"

type VehicleType struct {
	ID             uint64    `gorm:"primaryKey" json:"id"`
	Name           string    `gorm:"size:100;not null" json:"name"`
	NormalizedName string    `gorm:"->" json:"-"`
	Active         bool      `gorm:"not null;default:true" json:"active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (VehicleType) TableName() string { return "vehicle_types" }

type ServiceCategory struct {
	ID             uint64    `gorm:"primaryKey" json:"id"`
	Name           string    `gorm:"size:100;not null" json:"name"`
	NormalizedName string    `gorm:"->" json:"-"`
	Active         bool      `gorm:"not null;default:true" json:"active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (ServiceCategory) TableName() string { return "service_categories" }

type Service struct {
	ID             uint64          `gorm:"primaryKey" json:"id"`
	CategoryID     uint64          `gorm:"not null;index" json:"category_id"`
	Category       ServiceCategory `gorm:"foreignKey:CategoryID" json:"category"`
	Name           string          `gorm:"size:100;not null" json:"name"`
	NormalizedName string          `gorm:"->" json:"-"`
	Active         bool            `gorm:"not null;default:true" json:"active"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

func (Service) TableName() string { return "services" }

type ServicePrice struct {
	ID            uint64      `gorm:"primaryKey" json:"id"`
	ServiceID     uint64      `gorm:"not null;index" json:"service_id"`
	Service       Service     `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	VehicleTypeID uint64      `gorm:"not null;index" json:"vehicle_type_id"`
	VehicleType   VehicleType `gorm:"foreignKey:VehicleTypeID" json:"vehicle_type,omitempty"`
	Price         string      `gorm:"type:numeric(12,2);not null" json:"price"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

func (ServicePrice) TableName() string { return "service_prices" }
