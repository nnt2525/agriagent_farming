package agent

import (
	"context"
	"log"
	"time"

	"agriagent/backend/internal/config"
	"agriagent/backend/internal/llm"
	"agriagent/backend/internal/models"
	"agriagent/backend/internal/store"
)

// TriggerEvent describes why the agent is being asked to evaluate a device.
type TriggerEvent struct {
	DeviceID string
	Zone     string
	Source   string // "threshold" | "delta" | "cv" | "schedule" | "manual"

	// Optional context carried with the event; agent will look up latest
	// reading itself if Reading is nil (e.g. for schedule/manual triggers).
	Reading *models.SensorPayload

	CVResult     string
	CVConfidence float64

	// PrevSoilMoisture, when known by the caller (e.g. MQTT subscriber
	// comparing consecutive readings), is forwarded to the LLM as delta
	// context. Left nil for schedule/manual/cv triggers.
	PrevSoilMoisture *float64
}

// Agent is event-driven: callers push TriggerEvents onto a channel instead of
// the agent polling on a fixed interval. A background ticker only exists to
// emit the required "every 3h" schedule trigger per known device.
type Agent struct {
	cfg   config.Config
	store store.Store
	llm   llm.Client

	events chan TriggerEvent

	// ExecuteFunc, if set, is called for every auto-executed decision so
	// main.go can wire real relay actuation (e.g. publish an MQTT command)
	// without this package importing the MQTT client directly.
	ExecuteFunc func(models.AgentDecision)
}

func New(cfg config.Config, st store.Store, llmClient llm.Client) *Agent {
	return &Agent{
		cfg:    cfg,
		store:  st,
		llm:    llmClient,
		events: make(chan TriggerEvent, 64),
	}
}

// Trigger enqueues an evaluation. Non-blocking (buffered channel) so MQTT/API
// handlers never wait on the agent.
func (a *Agent) Trigger(ev TriggerEvent) {
	select {
	case a.events <- ev:
	default:
		log.Printf("agent: event queue full, dropping trigger for device=%s source=%s", ev.DeviceID, ev.Source)
	}
}

// Run consumes the event queue and drives the 3h schedule ticker. Blocks
// until ctx is cancelled — call with `go agent.Run(ctx)`.
func (a *Agent) Run(ctx context.Context) {
	ticker := time.NewTicker(a.cfg.ScheduleInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case ev := <-a.events:
			a.handle(ctx, ev)
		case <-ticker.C:
			a.scheduleSweep(ctx)
		}
	}
}

// scheduleSweep fires a "schedule" trigger for every known device so the
// 3h periodic re-evaluation requirement is met without polling per-second.
func (a *Agent) scheduleSweep(ctx context.Context) {
	devices, err := a.store.ListDevices(ctx)
	if err != nil {
		log.Printf("agent: schedule sweep list devices: %v", err)
		return
	}
	for _, d := range devices {
		if d.Type != "soil_sensor" {
			continue
		}
		a.handle(ctx, TriggerEvent{DeviceID: d.ID, Zone: d.Zone, Source: "schedule"})
	}
}

func (a *Agent) handle(ctx context.Context, ev TriggerEvent) {
	var reading models.SensorPayload

	if ev.Reading != nil {
		reading = *ev.Reading
	} else {
		latest, ok, err := a.store.LatestReadingForDevice(ctx, ev.DeviceID)
		if err != nil || !ok {
			log.Printf("agent: no reading available for device=%s, skipping", ev.DeviceID)
			return
		}
		reading = models.SensorPayload{
			DeviceID: latest.DeviceID, SoilMoisture: latest.SoilMoisture,
			Temperature: latest.Temperature, Humidity: latest.Humidity, Timestamp: latest.CreatedAt,
		}
	}

	in := llm.DecisionInput{
		DeviceID:      reading.DeviceID,
		Zone:          ev.Zone,
		SoilMoisture:  reading.SoilMoisture,
		Temperature:   reading.Temperature,
		Humidity:      reading.Humidity,
		PrevSoil:      ev.PrevSoilMoisture,
		CVResult:      ev.CVResult,
		CVConfidence:  ev.CVConfidence,
		TriggerSource: ev.Source,
	}

	decision, err := a.llm.Decide(ctx, in)
	if err != nil {
		log.Printf("agent: llm decide failed for device=%s: %v", ev.DeviceID, err)
		decision = models.LLMDecision{
			Action: "no_action", Reason: "LLM call failed: " + err.Error(),
			NeedHumanConfirm: true, Confidence: 0,
		}
	}

	status := "auto_executed"
	if decision.NeedHumanConfirm {
		status = "pending"
	}

	rec := models.AgentDecision{
		DeviceID:         ev.DeviceID,
		Action:           decision.Action,
		Reason:           decision.Reason,
		Confidence:       decision.Confidence,
		NeedHumanConfirm: decision.NeedHumanConfirm,
		Status:           status,
		TriggerSource:    ev.Source,
	}

	saved, err := a.store.InsertDecision(ctx, rec)
	if err != nil {
		log.Printf("agent: failed to persist decision for device=%s: %v", ev.DeviceID, err)
		return
	}

	if !decision.NeedHumanConfirm {
		a.execute(ctx, saved)
	} else {
		log.Printf("agent: decision %d for device=%s needs human confirmation (action=%s, confidence=%.2f)",
			saved.ID, ev.DeviceID, decision.Action, decision.Confidence)
	}
}

// execute performs the physical action (e.g. publish MQTT command to relay).
// Actuation is stubbed via a callback so this package has no MQTT dependency;
// wire ExecuteFunc in main.go once the relay control topic is defined.
func (a *Agent) execute(ctx context.Context, d models.AgentDecision) {
	log.Printf("agent: auto-executing decision %d device=%s action=%s reason=%q",
		d.ID, d.DeviceID, d.Action, d.Reason)
	if a.ExecuteFunc != nil {
		a.ExecuteFunc(d)
	}
}
