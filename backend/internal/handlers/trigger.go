package handlers

import (
	"github.com/gofiber/fiber/v2"

	"agriagent/backend/internal/agent"
)

type triggerRequest struct {
	DeviceID string `json:"device_id"`
}

// TriggerAgent lets the demo force an immediate agent evaluation for a
// device instead of waiting for the next sensor event or the 3h schedule.
func (h *Handler) TriggerAgent(c *fiber.Ctx) error {
	var req triggerRequest
	if err := c.BodyParser(&req); err != nil || req.DeviceID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "device_id is required")
	}

	zone := ""
	if devices, err := h.Store.ListDevices(c.Context()); err == nil {
		for _, d := range devices {
			if d.ID == req.DeviceID {
				zone = d.Zone
				break
			}
		}
	}

	h.Agent.Trigger(agent.TriggerEvent{
		DeviceID: req.DeviceID,
		Zone:     zone,
		Source:   "manual",
	})

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"status":  "triggered",
		"message": "agent evaluation queued, check /api/decisions shortly",
	})
}
