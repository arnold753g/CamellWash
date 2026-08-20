package models

import "time"

type Client struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	FullName     string    `gorm:"size:150;not null" json:"full_name"`
	CI           *string   `gorm:"size:30" json:"ci"`
	NormalizedCI *string   `gorm:"->" json:"-"`
	Phone        string    `gorm:"size:16;not null" json:"phone"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (Client) TableName() string { return "clients" }

type Vehicle struct {
	ID              uint64      `gorm:"primaryKey" json:"id"`
	ClientID        uint64      `gorm:"not null;index" json:"client_id"`
	Client          Client      `gorm:"foreignKey:ClientID" json:"-"`
	ClientName      string      `gorm:"-" json:"client_name,omitempty"`
	VehicleTypeID   uint64      `gorm:"not null;index" json:"vehicle_type_id"`
	VehicleType     VehicleType `gorm:"foreignKey:VehicleTypeID" json:"vehicle_type"`
	Plate           *string     `gorm:"size:20" json:"plate"`
	NormalizedPlate *string     `gorm:"->" json:"-"`
	Brand           string      `gorm:"size:100;not null" json:"brand"`
	Model           string      `gorm:"size:100;not null" json:"model"`
	Color           string      `gorm:"size:100;not null" json:"color"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

func (Vehicle) TableName() string { return "vehicles" }
