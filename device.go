package models

import "time"

// Device represents a physical device in the inventory system.
// This is the core domain object of our service.
type Device struct {
	ID        string    `db:"id"         json:"id"`
	IMEI      string    `db:"imei"       json:"imei"`
	Model     string    `db:"model"      json:"model"`
	Status    string    `db:"status"     json:"status"`
	Grade     string    `db:"grade"      json:"grade"`
	Price     float64   `db:"price"      json:"price"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// DeviceStatus represents the valid lifecycle states a device can be in.
// Using constants prevents typos and makes transitions easy to reason about.
type DeviceStatus string

const (
	StatusReceived DeviceStatus = "received"
	StatusTesting  DeviceStatus = "testing"
	StatusGraded   DeviceStatus = "graded"
	StatusSold     DeviceStatus = "sold"
)

// IsValidStatus checks if a given string is an allowed device status.
func IsValidStatus(s string) bool {
	switch DeviceStatus(s) {
	case StatusReceived, StatusTesting, StatusGraded, StatusSold:
		return true
	}
	return false
}

// ProcessResult is used to communicate the outcome of an async
// processing job back to the caller via a channel.
type ProcessResult struct {
	DeviceID string `json:"device_id"`
	NewGrade string `json:"new_grade"`
	Message  string `json:"message"`
	Error    error  `json:"error,omitempty"`
}

// CreateDeviceRequest is the expected JSON body for POST /devices
type CreateDeviceRequest struct {
	IMEI  string  `json:"imei"`
	Model string  `json:"model"`
	Price float64 `json:"price"`
}

// UpdateStatusRequest is the expected JSON body for PUT /devices/:id/status
type UpdateStatusRequest struct {
	Status string `json:"status"`
}
