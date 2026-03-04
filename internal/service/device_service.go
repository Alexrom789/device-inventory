package service

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/alexrom789/device-inventory/internal/models"
	"github.com/alexrom789/device-inventory/internal/repository"
)

// DeviceService contains the business logic for the inventory system.
// It sits between handlers (HTTP layer) and the repository (DB layer).
// This separation means business rules live in one place — not scattered
// across HTTP handlers or SQL queries.
type DeviceService struct {
	repo repository.DeviceRepository
}

// NewDeviceService is the constructor for DeviceService.
func NewDeviceService(repo repository.DeviceRepository) *DeviceService {
	return &DeviceService{repo: repo}
}

func (s *DeviceService) CreateDevice(req *models.CreateDeviceRequest) (*models.Device, error) {
	if req.IMEI == "" {
		return nil, fmt.Errorf("IMEI is required")
	}
	if req.Model == "" {
		return nil, fmt.Errorf("model is required")
	}

	device := &models.Device{
		IMEI:  req.IMEI,
		Model: req.Model,
		Price: req.Price,
	}

	return s.repo.Create(device)
}

func (s *DeviceService) GetDevice(id string) (*models.Device, error) {
	return s.repo.GetByID(id)
}

func (s *DeviceService) GetAllDevices() ([]*models.Device, error) {
	return s.repo.GetAll()
}

func (s *DeviceService) UpdateStatus(id string, req *models.UpdateStatusRequest) (*models.Device, error) {
	if !models.IsValidStatus(req.Status) {
		return nil, fmt.Errorf("invalid status: %s", req.Status)
	}

	return s.repo.UpdateStatus(id, req.Status)
}

// ProcessDevice is the star of the show for interview purposes.
//
// It demonstrates Go's concurrency model:
//   - We launch a goroutine to simulate async warehouse processing
//     (grading/testing a device takes time — we don't want to block the HTTP request)
//   - We use a channel to safely receive the result back from that goroutine
//   - We use a select with a timeout so the API never hangs indefinitely
//
// In a real system, you might not wait for the result at all and instead
// use a job queue. But this pattern shows you understand both goroutines AND channels.
func (s *DeviceService) ProcessDevice(id string) (*models.ProcessResult, error) {
	// First confirm the device exists
	device, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("device not found: %w", err)
	}

	// Mark the device as "testing" immediately (synchronous)
	_, err = s.repo.UpdateStatus(id, string(models.StatusTesting))
	if err != nil {
		return nil, fmt.Errorf("failed to start processing: %w", err)
	}

	// resultChan is a channel that will carry exactly one ProcessResult
	// from the goroutine back to this function.
	// Channels are typed — this one can only carry ProcessResult values.
	resultChan := make(chan models.ProcessResult, 1)

	// Launch the goroutine. The `go` keyword is all it takes.
	// This runs simulateGrading concurrently without blocking the current goroutine.
	// Think of it like a very cheap background thread — Go can run millions of these.
	go s.simulateGrading(device, resultChan)

	// select lets us wait on multiple channel operations simultaneously.
	// This is how we implement a timeout — we don't want to wait forever.
	select {
	case result := <-resultChan:
		// The goroutine finished and sent a result through the channel
		if result.Error != nil {
			return nil, result.Error
		}

		// Persist the grade to the database
		err = s.repo.UpdateGrade(id, result.NewGrade)
		if err != nil {
			return nil, err
		}

		return &result, nil

	case <-time.After(10 * time.Second):
		// If the goroutine hasn't responded in 10 seconds, we give up.
		// In production this might trigger a retry or dead-letter queue.
		return nil, fmt.Errorf("processing timed out for device %s", id)
	}
}

// simulateGrading mimics a warehouse testing workflow that takes real time.
// It runs inside a goroutine and communicates its result back via the channel.
//
// In production, this could be:
//   - Calling an external grading API
//   - Running diagnostic tests on the device
//   - Invoking ML model for cosmetic grade prediction
func (s *DeviceService) simulateGrading(device *models.Device, resultChan chan<- models.ProcessResult) {
	// Simulate variable processing time (1-3 seconds)
	processingTime := time.Duration(1+rand.Intn(3)) * time.Second
	time.Sleep(processingTime)

	// Simulate a grading outcome
	grades := []string{"A", "B", "C", "F"}
	grade := grades[rand.Intn(len(grades))]

	// Send the result back through the channel.
	// The <- operator here means "send into channel".
	resultChan <- models.ProcessResult{
		DeviceID: device.ID,
		NewGrade: grade,
		Message:  fmt.Sprintf("Device %s graded as %s after %v of testing", device.Model, grade, processingTime),
	}
}
