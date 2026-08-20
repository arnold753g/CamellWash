package seed

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"starter-backend/internal/models"
	"starter-backend/internal/security"

	"gorm.io/gorm"
)

func Development(db *gorm.DB, appEnv string, bcryptCost int) error {
	if appEnv == "production" {
		return errors.New("development seed is disabled in production")
	}
	if !strings.EqualFold(strings.TrimSpace(os.Getenv("SEED_DEVELOPMENT")), "true") {
		return errors.New("set SEED_DEVELOPMENT=true to confirm development seed")
	}
	email := strings.ToLower(strings.TrimSpace(os.Getenv("DEV_ADMIN_EMAIL")))
	password := os.Getenv("DEV_ADMIN_PASSWORD")
	if email == "" || password == "" {
		return errors.New("DEV_ADMIN_EMAIL and DEV_ADMIN_PASSWORD are required")
	}
	hash, err := security.HashPassword(password, bcryptCost)
	if err != nil {
		return fmt.Errorf("DEV_ADMIN_PASSWORD: %w", err)
	}
	user := models.User{Email: email, PasswordHash: hash, FirstName: "Administrador",
		LastName: "Principal", Role: models.RoleAdmin, Status: models.StatusActive}
	var existing models.User
	result := db.Where("email = ?", email).Limit(1).Find(&existing)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		if err := db.Create(&user).Error; err != nil {
			return fmt.Errorf("create initial administrator: %w", err)
		}
		return nil
	}
	if existing.Role != models.RoleAdmin {
		return errors.New("DEV_ADMIN_EMAIL belongs to a non-administrator account")
	}
	return nil
}
