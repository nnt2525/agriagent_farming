package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"agriagent/backend/internal/models"
)

const geminiEndpointFmt = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s"

// GeminiClient calls Google's Gemini API and asks it to return the decision
// strictly as JSON matching models.LLMDecision.
type GeminiClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewGeminiClient(apiKey, model string) *GeminiClient {
	return &GeminiClient{
		apiKey:     apiKey,
		model:      model,
		httpClient: &http.Client{Timeout: 20 * time.Second},
	}
}

type geminiRequest struct {
	Contents         []geminiContent        `json:"contents"`
	GenerationConfig geminiGenerationConfig `json:"generationConfig"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerationConfig struct {
	ResponseMimeType string `json:"responseMimeType"`
	Temperature      float64 `json:"temperature"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []geminiPart `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func (c *GeminiClient) Decide(ctx context.Context, in DecisionInput) (models.LLMDecision, error) {
	prompt := buildPrompt(in)

	reqBody := geminiRequest{
		Contents: []geminiContent{{Parts: []geminiPart{{Text: prompt}}}},
		GenerationConfig: geminiGenerationConfig{
			ResponseMimeType: "application/json",
			Temperature:      0.2,
		},
	}
	buf, err := json.Marshal(reqBody)
	if err != nil {
		return models.LLMDecision{}, fmt.Errorf("marshal gemini request: %w", err)
	}

	url := fmt.Sprintf(geminiEndpointFmt, c.model, c.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(buf))
	if err != nil {
		return models.LLMDecision{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return models.LLMDecision{}, fmt.Errorf("call gemini: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.LLMDecision{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return models.LLMDecision{}, fmt.Errorf("gemini returned %d: %s", resp.StatusCode, string(body))
	}

	var gr geminiResponse
	if err := json.Unmarshal(body, &gr); err != nil {
		return models.LLMDecision{}, fmt.Errorf("unmarshal gemini response: %w", err)
	}
	if len(gr.Candidates) == 0 || len(gr.Candidates[0].Content.Parts) == 0 {
		return models.LLMDecision{}, fmt.Errorf("gemini returned no candidates")
	}

	var decision models.LLMDecision
	text := gr.Candidates[0].Content.Parts[0].Text
	if err := json.Unmarshal([]byte(text), &decision); err != nil {
		return models.LLMDecision{}, fmt.Errorf("parse decision JSON %q: %w", text, err)
	}
	return decision, nil
}

func buildPrompt(in DecisionInput) string {
	prevStr := "unknown"
	if in.PrevSoil != nil {
		prevStr = fmt.Sprintf("%.1f%%", *in.PrevSoil)
	}
	cv := in.CVResult
	if cv == "" {
		cv = "no recent image"
	}

	return fmt.Sprintf(`You are the irrigation & crop-health decision agent for an organic tomato farm called AgriAgent.

Current sensor reading for device %s (zone %s):
- soil_moisture: %.1f%%
- temperature: %.1fC
- humidity: %.1f%%
- previous soil_moisture: %s
- leaf CV result: %s (confidence %.2f)
- trigger reason: %s

Decide the correct action for the irrigation relay in this zone.
Respond with ONLY a JSON object, no markdown, matching exactly this schema:
{
  "action": "water_on" | "water_off" | "no_action" | "alert",
  "reason": "short explanation grounded in the numbers above",
  "need_human_confirm": true | false,
  "confidence": 0.0-1.0
}

Guidance:
- Organic tomato plants prefer soil moisture roughly 30-60%%.
- If confidence in the correct action is high (>=0.85), set need_human_confirm to false so it can auto-execute.
- If data is ambiguous, conflicting, or suggests a sensor/CV anomaly, set need_human_confirm to true.
- Use "alert" when the leaf CV result suggests disease/pest damage.`,
		in.DeviceID, in.Zone, in.SoilMoisture, in.Temperature, in.Humidity, prevStr, cv, in.CVConfidence, in.TriggerSource)
}
