package handlers

import "github.com/gofiber/fiber/v2"

func (h *Handler) RecentImages(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 20)
	images, err := h.Store.RecentImages(c.Context(), limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(images)
}
