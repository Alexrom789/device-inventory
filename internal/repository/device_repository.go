package repository

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/alexrom789/device-inventory/internal/models"
)

// DeviceRepository defines the interface (contract) for all database
// operations on devices. Using an interface here means we could swap
// out Postgres for any other DB without changing the service layer.
type DeviceRepository interface {
	Create(device *models.Device) (*models.Device, error)
	GetByID(id string) (*models.Device, error)
	GetAll() ([]*models.Device, error)
	UpdateStatus(id string, status string) (*models.Device, error)
	UpdateGrade(id string, grade string) error
}

// postgresDeviceRepo is the concrete Postgres implementation of DeviceRepository.
// It is unexported — callers use the interface, not this struct directly.
type postgresDeviceRepo struct {
	db *sqlx.DB
}

// NewDeviceRepository is a constructor that returns the interface type.
// This pattern (returning an interface from a constructor) is idiomatic Go.
func NewDeviceRepository(db *sqlx.DB) DeviceRepository {
	return &postgresDeviceRepo{db: db}
}

// Create inserts a new device record into Postgres.
// We generate the UUID here in application code rather than the DB,
// which makes the ID available immediately without a roundtrip query.
func (r *postgresDeviceRepo) Create(req *models.Device) (*models.Device, error) {
	device := &models.Device{
		ID:        uuid.New().String(),
		IMEI:      req.IMEI,
		Model:     req.Model,
		Status:    string(models.StatusReceived),
		Grade:     "ungraded",
		Price:     req.Price,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	query := `
		INSERT INTO devices (id, imei, model, status, grade, price, created_at, updated_at)
		VALUES (:id, :imei, :model, :status, :grade, :price, :created_at, :updated_at)
	`

	_, err := r.db.NamedExec(query, device)
	if err != nil {
		return nil, fmt.Errorf("failed to create device: %w", err)
	}

	return device, nil
}

// GetByID fetches a single device by its UUID primary key.
func (r *postgresDeviceRepo) GetByID(id string) (*models.Device, error) {
	var device models.Device

	query := `SELECT * FROM devices WHERE id = $1`
	err := r.db.Get(&device, query, id)
	if err != nil {
		return nil, fmt.Errorf("device not found: %w", err)
	}

	return &device, nil
}

// GetAll returns every device in the inventory, most recently updated first.
func (r *postgresDeviceRepo) GetAll() ([]*models.Device, error) {
	var devices []*models.Device

	query := `SELECT * FROM devices ORDER BY updated_at DESC`
	err := r.db.Select(&devices, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch devices: %w", err)
	}

	return devices, nil
}

// UpdateStatus transitions a device to a new lifecycle status.
// It returns the updated device so the handler can respond with fresh data.
func (r *postgresDeviceRepo) UpdateStatus(id string, status string) (*models.Device, error) {
	query := `
		UPDATE devices
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.Exec(query, status, time.Now(), id)
	if err != nil {
		return nil, fmt.Errorf("failed to update status: %w", err)
	}

	return r.GetByID(id)
}

// UpdateGrade sets the grade field after async processing completes.
func (r *postgresDeviceRepo) UpdateGrade(id string, grade string) error {
	query := `
		UPDATE devices
		SET grade = $1, status = $2, updated_at = $3
		WHERE id = $4
	`

	_, err := r.db.Exec(query, grade, string(models.StatusGraded), time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update grade: %w", err)
	}

	return nil
}
