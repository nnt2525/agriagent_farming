package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) ListDecisions(c *fiber.Ctx) error {
	decisions, err := h.Store.ListDecisions(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(decisions)
}

type confirmRequest struct {
	Approve     bool   `json:"approve"`
	ConfirmedBy string `json:"confirmed_by"`
}

func (h *Handler) ConfirmDecision(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid decision id")
	}

	var req confirmRequest
	// Body is optional; default to approve=true so a bare POST from the demo
	// "confirm" button just works.
	req.Approve = true
	req.ConfirmedBy = "admin"
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid body")
		}
	}
	if req.ConfirmedBy == "" {
		req.ConfirmedBy = "admin"
	}

	updated, err := h.Store.ConfirmDecision(c.Context(), id, req.ConfirmedBy, req.Approve)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return c.JSON(updated)
}
