package models

import "time"

const (
	OrderPending        = "PENDING"
	OrderInProgress     = "IN_PROGRESS"
	OrderQualityControl = "QUALITY_CONTROL"
	OrderFinished       = "FINISHED"
	OrderDelivered      = "DELIVERED"
	OrderCancelled      = "CANCELLED"
)

type Order struct {
	ID                  uint64         `gorm:"primaryKey" json:"id"`
	VehicleID           uint64         `json:"vehicle_id"`
	Vehicle             Vehicle        `gorm:"foreignKey:VehicleID" json:"vehicle"`
	RegisteredByUserID  uint64         `json:"registered_by_user_id"`
	RegisteredBy        User           `gorm:"foreignKey:RegisteredByUserID" json:"registered_by"`
	ReceivedAt          time.Time      `json:"received_at"`
	EstimatedDeliveryAt time.Time      `json:"estimated_delivery_at"`
	Status              string         `json:"status"`
	Total               string         `gorm:"type:numeric(12,2)" json:"total"`
	Version             uint64         `json:"version"`
	Services            []OrderService `gorm:"foreignKey:OrderID" json:"services"`
	Staff               []User         `gorm:"many2many:order_staff;joinForeignKey:OrderID;joinReferences:UserID" json:"staff"`
	History             []OrderHistory `gorm:"foreignKey:OrderID" json:"history"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
}

func (Order) TableName() string { return "orders" }

type OrderService struct {
	ID                  uint64    `gorm:"primaryKey" json:"id"`
	OrderID             uint64    `json:"order_id"`
	ServiceID           uint64    `json:"service_id"`
	ServiceNameSnapshot string    `json:"service_name_snapshot"`
	PriceSnapshot       string    `gorm:"type:numeric(12,2)" json:"price_snapshot"`
	CreatedAt           time.Time `json:"created_at"`
}

func (OrderService) TableName() string { return "order_services" }

type OrderStaff struct {
	OrderID    uint64    `json:"order_id"`
	UserID     uint64    `json:"user_id"`
	AssignedAt time.Time `json:"assigned_at"`
}

func (OrderStaff) TableName() string { return "order_staff" }

type OrderHistory struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	OrderID     uint64    `json:"order_id"`
	FromStatus  *string   `json:"from_status"`
	ToStatus    string    `json:"to_status"`
	ChangedAt   time.Time `json:"changed_at"`
	ActorUserID uint64    `json:"actor_user_id"`
	Actor       User      `gorm:"foreignKey:ActorUserID" json:"actor"`
	Reason      *string   `json:"reason"`
}

func (OrderHistory) TableName() string { return "order_status_history" }
