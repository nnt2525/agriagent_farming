package llm

import (
	"context"

	"agriagent/backend/internal/models"
)

// MockClient returns deterministic, rule-based decisions without calling any
// external API. Used automatically when GEMINI_API_KEY is not set, so the
// agent pipeline (trigger -> decide -> store -> maybe auto-execute) can be
// demoed end-to-end offline.
type MockClient struct{}

func NewMockClient() *MockClient { return &MockClient{} }

func (m *MockClient) Decide(ctx context.Context, in DecisionInput) (models.LLMDecision, error) {
	if in.CVResult != "" && in.CVResult != "healthy" {
		conf := in.CVConfidence
		return models.LLMDecision{
			Action:           "alert",
			Reason:           "CV flagged possible leaf issue (" + in.CVResult + "); recommend visual inspection before acting.",
			NeedHumanConfirm: conf < 0.85,
			Confidence:       conf,
		}, nil
	}

	if in.SoilMoisture < 30 {
		return models.LLMDecision{
			Action:           "water_on",
			Reason:           "Soil moisture below 30% threshold for organic tomato zone; irrigation recommended.",
			NeedHumanConfirm: false,
			Confidence:       0.9,
		}, nil
	}

	if in.PrevSoil != nil {
		delta := *in.PrevSoil - in.SoilMoisture
		if delta > 10 {
			return models.LLMDecision{
				Action:           "water_on",
				Reason:           "Sudden soil moisture drop >10% detected; could be sensor fault or fast drainage, confirm before irrigating.",
				NeedHumanConfirm: true,
				Confidence:       0.6,
			}, nil
		}
	}

	return models.LLMDecision{
		Action:           "no_action",
		Reason:           "Soil moisture, temperature and humidity within healthy range; no intervention needed.",
		NeedHumanConfirm: false,
		Confidence:       0.85,
	}, nil
}
