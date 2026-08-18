package handlers

import "github.com/gofiber/fiber/v2"

func (h *Handler) LatestReadings(c *fiber.Ctx) error {
	readings, err := h.Store.LatestReadings(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(readings)
}

func (h *Handler) ReadingsRange(c *fiber.Ctx) error {
	rangeKey := c.Query("range", "day")
	switch rangeKey {
	case "day", "month", "year":
	default:
		return fiber.NewError(fiber.StatusBadRequest, "range must be one of day|month|year")
	}
	readings, err := h.Store.ReadingsRange(c.Context(), rangeKey)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(readings)
}
