package models

import "time"

type Session struct {
	ID             string     `gorm:"primaryKey;size:64"`
	UserID         uint64     `gorm:"not null;index"`
	TokenHash      string     `gorm:"size:64;not null;uniqueIndex"`
	LastActivityAt time.Time  `gorm:"not null;index"`
	ExpiresAt      time.Time  `gorm:"not null;index"`
	UserAgent      string     `gorm:"size:512"`
	IPAddress      string     `gorm:"size:64"`
	RevokedAt      *time.Time `gorm:"index"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	User           User `gorm:"foreignKey:UserID"`
}

func (Session) TableName() string { return "auth_sessions" }
