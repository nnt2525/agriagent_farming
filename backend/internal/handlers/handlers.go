package handlers

import (
	"github.com/gofiber/fiber/v2"

	"agriagent/backend/internal/agent"
	"agriagent/backend/internal/store"
)

type Handler struct {
	Store store.Store
	Agent *agent.Agent
}

func New(st store.Store, ag *agent.Agent) *Handler {
	return &Handler{Store: st, Agent: ag}
}

func (h *Handler) Register(api fiber.Router) {
	api.Get("/readings/latest", h.LatestReadings)
	api.Get("/readings", h.ReadingsRange)
	api.Get("/devices", h.ListDevices)
	api.Get("/images/recent", h.RecentImages)
	api.Get("/decisions", h.ListDecisions)
	api.Post("/decisions/:id/confirm", h.ConfirmDecision)
	api.Post("/agent/trigger", h.TriggerAgent)
}
