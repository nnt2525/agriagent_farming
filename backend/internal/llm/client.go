package llm

import (
	"context"

	"agriagent/backend/internal/models"
)

// DecisionInput bundles the context sent to the LLM for one agent evaluation.
type DecisionInput struct {
	DeviceID      string
	Zone          string
	SoilMoisture  float64
	Temperature   float64
	Humidity      float64
	PrevSoil      *float64 // previous reading, for delta context
	CVResult      string   // e.g. "healthy", "yellowing" — empty if no recent image
	CVConfidence  float64
	TriggerSource string // "threshold" | "delta" | "cv" | "schedule" | "manual"
}

// Client is implemented by both the real Gemini client and the offline mock,
// so the agent package never needs to know which one is active.
type Client interface {
	Decide(ctx context.Context, in DecisionInput) (models.LLMDecision, error)
}
