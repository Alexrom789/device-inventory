package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/alexrom789/device-inventory/internal/models"
	"github.com/alexrom789/device-inventory/internal/service"
)

// DeviceHandler holds a reference to the service layer.
// Handlers are responsible ONLY for:
//   1. Parsing the HTTP request
//   2. Calling the appropriate service method
//   3. Writing the HTTP response
//
// Business logic does NOT belong here — it lives in the service layer.
type DeviceHandler struct {
	service *service.DeviceService
}

// NewDeviceHandler constructs a DeviceHandler with its dependency injected.
func NewDeviceHandler(service *service.DeviceService) *DeviceHandler {
	return &DeviceHandler{service: service}
}

// RegisterRoutes wires all device endpoints to their handler methods.
// Grouping routes like this keeps main.go clean.
func (h *DeviceHandler) RegisterRoutes(app *fiber.App) {
	devices := app.Group("/devices")

	devices.Post("/", h.CreateDevice)
	devices.Get("/", h.GetAllDevices)
	devices.Get("/:id", h.GetDevice)
	devices.Put("/:id/status", h.UpdateStatus)
	devices.Post("/:id/process", h.ProcessDevice)
}

// CreateDevice handles POST /devices
func (h *DeviceHandler) CreateDevice(c *fiber.Ctx) error {
	var req models.CreateDeviceRequest

	// BodyParser deserializes the JSON body into our request struct
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	device, err := h.service.CreateDevice(&req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(device)
}

// GetDevice handles GET /devices/:id
func (h *DeviceHandler) GetDevice(c *fiber.Ctx) error {
	id := c.Params("id")

	device, err := h.service.GetDevice(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "device not found",
		})
	}

	return c.JSON(device)
}

// GetAllDevices handles GET /devices
func (h *DeviceHandler) GetAllDevices(c *fiber.Ctx) error {
	devices, err := h.service.GetAllDevices()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch devices",
		})
	}

	return c.JSON(fiber.Map{
		"count":   len(devices),
		"devices": devices,
	})
}

// UpdateStatus handles PUT /devices/:id/status
func (h *DeviceHandler) UpdateStatus(c *fiber.Ctx) error {
	id := c.Params("id")

	var req models.UpdateStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	device, err := h.service.UpdateStatus(id, &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(device)
}

// ProcessDevice handles POST /devices/:id/process
// This endpoint demonstrates Go's concurrency — it launches a goroutine
// to simulate async grading and waits for the result via a channel.
// The request blocks until grading completes (or times out at 10s).
func (h *DeviceHandler) ProcessDevice(c *fiber.Ctx) error {
	id := c.Params("id")

	result, err := h.service.ProcessDevice(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(result)
}
