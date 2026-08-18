package handlers

import (
	"github.com/gofiber/fiber/v2"

	"agriagent/backend/internal/models"
)

type devicesResponse struct {
	Active   int             `json:"active"`
	Inactive int             `json:"inactive"`
	Devices  []models.Device `json:"devices"`
}

func (h *Handler) ListDevices(c *fiber.Ctx) error {
	devices, err := h.Store.ListDevices(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	resp := devicesResponse{Devices: devices}
	for _, d := range devices {
		if d.Status == "active" {
			resp.Active++
		} else {
			resp.Inactive++
		}
	}
	return c.JSON(resp)
}
