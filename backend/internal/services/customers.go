package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"starter-backend/internal/models"

	"gorm.io/gorm"
)

const MissingCIDisplay = "00000000"

var (
	ErrClientNotFound         = errors.New("client not found")
	ErrVehicleNotFound        = errors.New("vehicle not found")
	ErrCIExists               = errors.New("client ci already exists")
	ErrPlateExists            = errors.New("vehicle plate already exists")
	ErrVehicleTypeUnavailable = errors.New("vehicle type is unavailable")
)

type FieldError struct {
	Field   string
	Message string
}

func (e FieldError) Error() string { return e.Message }

type CustomerService struct {
	db *gorm.DB
}

type ClientInput struct {
	FullName string
	CI       *string
	Phone    string
}

type VehicleInput struct {
	VehicleTypeID uint64
	Plate         *string
	Brand         string
	Model         string
	Color         string
}

type ClientDetail struct {
	ID           uint64    `json:"id"`
	FullName     string    `json:"full_name"`
	CI           string    `json:"ci"`
	CIProvided   bool      `json:"ci_provided"`
	Phone        string    `json:"phone"`
	VehicleCount int64     `json:"vehicle_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ClientListItem struct {
	ID           uint64 `json:"id"`
	FullName     string `json:"full_name"`
	CI           string `json:"ci"`
	CIProvided   bool   `json:"ci_provided"`
	Phone        string `json:"phone"`
	VehicleCount int64  `json:"vehicle_count"`
}

type PageData struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}

func NewCustomerService(db *gorm.DB) *CustomerService { return &CustomerService{db: db} }

func NormalizeClientName(value string) (string, error) {
	value = strings.Join(strings.Fields(value), " ")
	if utf8.RuneCountInString(value) < 2 || utf8.RuneCountInString(value) > 150 {
		return "", FieldError{Field: "full_name", Message: "Name must contain between 2 and 150 characters"}
	}
	return value, nil
}

func NormalizeCI(value *string) (*string, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil, nil
	}
	display := strings.ToUpper(strings.Join(strings.Fields(*value), " "))
	var normalized strings.Builder
	for _, character := range display {
		switch {
		case isASCIIAlphaNumeric(character):
			normalized.WriteRune(character)
		case character == ' ', character == '-':
		default:
			return nil, FieldError{Field: "ci", Message: "CI may contain only letters, numbers, spaces or hyphens"}
		}
	}
	canonical := normalized.String()
	if len(canonical) < 4 || len(canonical) > 20 {
		return nil, FieldError{Field: "ci", Message: "CI must contain between 4 and 20 letters or numbers"}
	}
	if canonical == MissingCIDisplay {
		return nil, FieldError{Field: "ci", Message: "00000000 is reserved for clients who did not provide a CI"}
	}
	return &display, nil
}

func NormalizePhone(value string) (string, error) {
	value = strings.TrimSpace(value)
	var normalized strings.Builder
	for index, character := range value {
		switch {
		case character >= '0' && character <= '9':
			normalized.WriteRune(character)
		case character == '+' && index == 0:
			normalized.WriteRune(character)
		case character == ' ', character == '-', character == '(', character == ')':
		default:
			return "", FieldError{Field: "phone", Message: "Phone contains unsupported characters"}
		}
	}
	phone := normalized.String()
	digits := strings.TrimPrefix(phone, "+")
	if len(digits) < 7 || len(digits) > 15 {
		return "", FieldError{Field: "phone", Message: "Phone must contain between 7 and 15 digits"}
	}
	return phone, nil
}

func NormalizePlate(value *string) (*string, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil, nil
	}
	var normalized strings.Builder
	for _, character := range strings.ToUpper(*value) {
		switch {
		case isASCIIAlphaNumeric(character):
			normalized.WriteRune(character)
		case character == ' ', character == '-':
		default:
			return nil, FieldError{Field: "plate", Message: "Plate may contain only letters, numbers, spaces or hyphens"}
		}
	}
	plate := normalized.String()
	if len(plate) < 3 || len(plate) > 15 {
		return nil, FieldError{Field: "plate", Message: "Plate must contain between 3 and 15 letters or numbers"}
	}
	return &plate, nil
}

func NormalizeVehicleText(field, value string) (string, error) {
	value = strings.Join(strings.Fields(value), " ")
	if utf8.RuneCountInString(value) > 100 {
		return "", FieldError{Field: field, Message: "Use at most 100 characters"}
	}
	return value, nil
}

func ClientCIForDisplay(ci *string) (string, bool) {
	if ci == nil || strings.TrimSpace(*ci) == "" {
		return MissingCIDisplay, false
	}
	return *ci, true
}

func MaskPersonalValue(value string) string {
	runes := []rune(value)
	if len(runes) <= 4 {
		return strings.Repeat("•", len(runes))
	}
	return strings.Repeat("•", len(runes)-4) + string(runes[len(runes)-4:])
}

func (s *CustomerService) CreateClient(ctx context.Context, input ClientInput) (ClientDetail, error) {
	client, err := normalizedClient(input)
	if err != nil {
		return ClientDetail{}, err
	}
	if err := s.db.WithContext(ctx).Create(&client).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ClientDetail{}, ErrCIExists
		}
		return ClientDetail{}, fmt.Errorf("create client: %w", err)
	}
	return clientDetail(client, 0), nil
}

func (s *CustomerService) UpdateClient(ctx context.Context, id uint64, input ClientInput) (ClientDetail, error) {
	client, err := normalizedClient(input)
	if err != nil {
		return ClientDetail{}, err
	}
	result := s.db.WithContext(ctx).Model(&models.Client{}).Where("id = ?", id).Updates(map[string]any{
		"full_name": client.FullName, "ci": client.CI, "phone": client.Phone,
	})
	if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
		return ClientDetail{}, ErrCIExists
	}
	if result.Error != nil {
		return ClientDetail{}, fmt.Errorf("update client: %w", result.Error)
	}
	if result.RowsAffected != 1 {
		return ClientDetail{}, ErrClientNotFound
	}
	return s.GetClient(ctx, id)
}

func normalizedClient(input ClientInput) (models.Client, error) {
	name, err := NormalizeClientName(input.FullName)
	if err != nil {
		return models.Client{}, err
	}
	ci, err := NormalizeCI(input.CI)
	if err != nil {
		return models.Client{}, err
	}
	phone, err := NormalizePhone(input.Phone)
	if err != nil {
		return models.Client{}, err
	}
	return models.Client{FullName: name, CI: ci, Phone: phone}, nil
}

func (s *CustomerService) GetClient(ctx context.Context, id uint64) (ClientDetail, error) {
	var client models.Client
	if err := s.db.WithContext(ctx).First(&client, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ClientDetail{}, ErrClientNotFound
		}
		return ClientDetail{}, fmt.Errorf("load client: %w", err)
	}
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.Vehicle{}).Where("client_id = ?", id).Count(&count).Error; err != nil {
		return ClientDetail{}, fmt.Errorf("count client vehicles: %w", err)
	}
	return clientDetail(client, count), nil
}

func clientDetail(client models.Client, vehicleCount int64) ClientDetail {
	ci, provided := ClientCIForDisplay(client.CI)
	return ClientDetail{ID: client.ID, FullName: client.FullName, CI: ci, CIProvided: provided, Phone: client.Phone,
		VehicleCount: vehicleCount, CreatedAt: client.CreatedAt, UpdatedAt: client.UpdatedAt}
}

func (s *CustomerService) ListClients(ctx context.Context, search string, page, limit int) ([]ClientListItem, PageData, error) {
	query := s.db.WithContext(ctx).Model(&models.Client{})
	search = strings.TrimSpace(search)
	if search != "" {
		conditions := []string{"LOWER(full_name) LIKE ? ESCAPE '!'"}
		arguments := []any{"%" + escapeSearchLike(strings.ToLower(search)) + "%"}
		if identifier := normalizeSearchIdentifier(search); identifier != "" {
			conditions = append(conditions, "COALESCE(normalized_ci, '') LIKE ? ESCAPE '!'")
			arguments = append(arguments, "%"+escapeSearchLike(identifier)+"%")
		}
		if phone := normalizeSearchPhone(search); phone != "" {
			conditions = append(conditions, "phone LIKE ? ESCAPE '!'")
			arguments = append(arguments, "%"+escapeSearchLike(phone)+"%")
		}
		query = query.Where(strings.Join(conditions, " OR "), arguments...)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, PageData{}, fmt.Errorf("count clients: %w", err)
	}
	type row struct {
		models.Client
		VehicleCount int64
	}
	var rows []row
	if err := query.Select("clients.*, (SELECT COUNT(*) FROM vehicles WHERE vehicles.client_id = clients.id) AS vehicle_count").
		Order("LOWER(full_name), id").Offset((page - 1) * limit).Limit(limit).Scan(&rows).Error; err != nil {
		return nil, PageData{}, fmt.Errorf("list clients: %w", err)
	}
	items := make([]ClientListItem, len(rows))
	for index, item := range rows {
		ci, provided := ClientCIForDisplay(item.CI)
		if provided {
			ci = MaskPersonalValue(ci)
		}
		items[index] = ClientListItem{ID: item.ID, FullName: item.FullName, CI: ci, CIProvided: provided,
			Phone: MaskPersonalValue(item.Phone), VehicleCount: item.VehicleCount}
	}
	return items, pageData(page, limit, total), nil
}

func normalizeSearchIdentifier(value string) string {
	var output strings.Builder
	for _, character := range strings.ToUpper(value) {
		if isASCIIAlphaNumeric(character) {
			output.WriteRune(character)
		}
	}
	return output.String()
}

func normalizeSearchPhone(value string) string {
	var output strings.Builder
	for index, character := range value {
		if character >= '0' && character <= '9' || character == '+' && index == 0 {
			output.WriteRune(character)
		}
	}
	return output.String()
}

func (s *CustomerService) CreateVehicle(ctx context.Context, clientID uint64, input VehicleInput) (models.Vehicle, error) {
	if err := s.ensureClient(ctx, clientID); err != nil {
		return models.Vehicle{}, err
	}
	vehicle, err := s.normalizedVehicle(ctx, 0, clientID, input)
	if err != nil {
		return models.Vehicle{}, err
	}
	if err := s.db.WithContext(ctx).Create(&vehicle).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return models.Vehicle{}, ErrPlateExists
		}
		return models.Vehicle{}, fmt.Errorf("create vehicle: %w", err)
	}
	return s.GetVehicle(ctx, vehicle.ID)
}

func (s *CustomerService) UpdateVehicle(ctx context.Context, id uint64, input VehicleInput) (models.Vehicle, error) {
	var current models.Vehicle
	if err := s.db.WithContext(ctx).First(&current, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Vehicle{}, ErrVehicleNotFound
		}
		return models.Vehicle{}, fmt.Errorf("load vehicle: %w", err)
	}
	vehicle, err := s.normalizedVehicle(ctx, current.VehicleTypeID, current.ClientID, input)
	if err != nil {
		return models.Vehicle{}, err
	}
	result := s.db.WithContext(ctx).Model(&models.Vehicle{}).Where("id = ?", id).Updates(map[string]any{
		"vehicle_type_id": vehicle.VehicleTypeID, "plate": vehicle.Plate, "brand": vehicle.Brand, "model": vehicle.Model, "color": vehicle.Color,
	})
	if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
		return models.Vehicle{}, ErrPlateExists
	}
	if result.Error != nil {
		return models.Vehicle{}, fmt.Errorf("update vehicle: %w", result.Error)
	}
	if result.RowsAffected != 1 {
		return models.Vehicle{}, ErrVehicleNotFound
	}
	return s.GetVehicle(ctx, id)
}

func (s *CustomerService) normalizedVehicle(ctx context.Context, currentTypeID, clientID uint64, input VehicleInput) (models.Vehicle, error) {
	if input.VehicleTypeID == 0 {
		return models.Vehicle{}, FieldError{Field: "vehicle_type_id", Message: "Select a vehicle type"}
	}
	if input.VehicleTypeID != currentTypeID {
		var count int64
		if err := s.db.WithContext(ctx).Model(&models.VehicleType{}).Where("id = ? AND active = TRUE", input.VehicleTypeID).Count(&count).Error; err != nil {
			return models.Vehicle{}, fmt.Errorf("check vehicle type: %w", err)
		}
		if count != 1 {
			return models.Vehicle{}, ErrVehicleTypeUnavailable
		}
	}
	plate, err := NormalizePlate(input.Plate)
	if err != nil {
		return models.Vehicle{}, err
	}
	brand, err := NormalizeVehicleText("brand", input.Brand)
	if err != nil {
		return models.Vehicle{}, err
	}
	model, err := NormalizeVehicleText("model", input.Model)
	if err != nil {
		return models.Vehicle{}, err
	}
	color, err := NormalizeVehicleText("color", input.Color)
	if err != nil {
		return models.Vehicle{}, err
	}
	return models.Vehicle{ClientID: clientID, VehicleTypeID: input.VehicleTypeID, Plate: plate, Brand: brand, Model: model, Color: color}, nil
}

func (s *CustomerService) GetVehicle(ctx context.Context, id uint64) (models.Vehicle, error) {
	var vehicle models.Vehicle
	if err := s.db.WithContext(ctx).Preload("VehicleType").First(&vehicle, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Vehicle{}, ErrVehicleNotFound
		}
		return models.Vehicle{}, fmt.Errorf("load vehicle: %w", err)
	}
	return vehicle, nil
}

func (s *CustomerService) ListVehicles(ctx context.Context, clientID uint64, search string, page, limit int) ([]models.Vehicle, PageData, error) {
	if err := s.ensureClient(ctx, clientID); err != nil {
		return nil, PageData{}, err
	}
	query := s.db.WithContext(ctx).Model(&models.Vehicle{}).Where("vehicles.client_id = ?", clientID)
	if search = strings.TrimSpace(search); search != "" {
		like := "%" + escapeSearchLike(strings.ToLower(search)) + "%"
		conditions := []string{
			"LOWER(vehicles.brand) LIKE ? ESCAPE '!'",
			"LOWER(vehicles.model) LIKE ? ESCAPE '!'",
			"LOWER(vehicles.color) LIKE ? ESCAPE '!'",
			"LOWER(vt_search.name) LIKE ? ESCAPE '!'",
		}
		arguments := []any{like, like, like, like}
		if plate := normalizeSearchIdentifier(search); plate != "" {
			conditions = append(conditions, "COALESCE(vehicles.normalized_plate, '') LIKE ? ESCAPE '!'")
			arguments = append(arguments, "%"+escapeSearchLike(plate)+"%")
		}
		query = query.Joins("JOIN vehicle_types vt_search ON vt_search.id = vehicles.vehicle_type_id").Where(strings.Join(conditions, " OR "), arguments...)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, PageData{}, fmt.Errorf("count vehicles: %w", err)
	}
	var items []models.Vehicle
	if err := query.Preload("VehicleType").Order("vehicles.created_at DESC, vehicles.id DESC").Offset((page - 1) * limit).Limit(limit).Find(&items).Error; err != nil {
		return nil, PageData{}, fmt.Errorf("list vehicles: %w", err)
	}
	return items, pageData(page, limit, total), nil
}

func (s *CustomerService) ensureClient(ctx context.Context, id uint64) error {
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.Client{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return fmt.Errorf("check client: %w", err)
	}
	if count != 1 {
		return ErrClientNotFound
	}
	return nil
}

func pageData(page, limit int, total int64) PageData {
	return PageData{Page: page, Limit: limit, Total: total, TotalPages: (total + int64(limit) - 1) / int64(limit)}
}

func isASCIIAlphaNumeric(character rune) bool {
	return character >= 'A' && character <= 'Z' || character >= '0' && character <= '9'
}

func escapeSearchLike(value string) string {
	return strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(value)
}
