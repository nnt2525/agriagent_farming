package mqtt

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"

	"agriagent/backend/internal/agent"
	"agriagent/backend/internal/config"
	"agriagent/backend/internal/models"
	"agriagent/backend/internal/store"
)

// Subscriber connects to the MQTT broker (HiveMQ Cloud or any TLS broker),
// listens for ESP32 telemetry, persists readings, and forwards event-driven
// triggers (threshold breach / sudden delta) to the Agent.
type Subscriber struct {
	cfg   config.Config
	store store.Store
	agent *agent.Agent
	client paho.Client
}

func New(cfg config.Config, st store.Store, ag *agent.Agent) *Subscriber {
	return &Subscriber{cfg: cfg, store: st, agent: ag}
}

// Connect dials the broker and subscribes. Returns an error immediately if
// MQTT_BROKER_URL is unset so callers can skip MQTT entirely in mock-only
// local dev.
func (s *Subscriber) Connect() error {
	opts := paho.NewClientOptions().
		AddBroker(s.cfg.MQTTBrokerURL).
		SetClientID(s.cfg.MQTTClientID).
		SetUsername(s.cfg.MQTTUsername).
		SetPassword(s.cfg.MQTTPassword).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectTimeout(10 * time.Second)

	opts.OnConnect = func(c paho.Client) {
		log.Printf("mqtt: connected to %s, subscribing to %s", s.cfg.MQTTBrokerURL, s.cfg.MQTTTopic)
		if token := c.Subscribe(s.cfg.MQTTTopic, 1, s.onMessage); token.Wait() && token.Error() != nil {
			log.Printf("mqtt: subscribe error: %v", token.Error())
		}
	}
	opts.OnConnectionLost = func(c paho.Client, err error) {
		log.Printf("mqtt: connection lost: %v", err)
	}

	s.client = paho.NewClient(opts)
	token := s.client.Connect()
	token.Wait()
	return token.Error()
}

func (s *Subscriber) Disconnect() {
	if s.client != nil && s.client.IsConnected() {
		s.client.Disconnect(250)
	}
}

func (s *Subscriber) onMessage(_ paho.Client, msg paho.Message) {
	var payload models.SensorPayload
	if err := json.Unmarshal(msg.Payload(), &payload); err != nil {
		log.Printf("mqtt: invalid payload on topic %s: %v", msg.Topic(), err)
		return
	}
	if payload.DeviceID == "" {
		log.Printf("mqtt: payload missing device_id on topic %s", msg.Topic())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Capture the previous reading BEFORE inserting the new one so we can
	// compute the sudden-change delta trigger.
	var prevSoil *float64
	if prev, ok, err := s.store.LatestReadingForDevice(ctx, payload.DeviceID); err == nil && ok {
		v := prev.SoilMoisture
		prevSoil = &v
	}

	if _, err := s.store.InsertReading(ctx, payload); err != nil {
		log.Printf("mqtt: failed to insert reading for device=%s: %v", payload.DeviceID, err)
		return
	}
	if err := s.store.TouchDevice(ctx, payload.DeviceID); err != nil {
		log.Printf("mqtt: failed to touch device=%s: %v", payload.DeviceID, err)
	}

	s.evaluateTriggers(payload, prevSoil)
}

func (s *Subscriber) evaluateTriggers(payload models.SensorPayload, prevSoil *float64) {
	if payload.SoilMoisture < s.cfg.SoilMoistureLowPct {
		s.agent.Trigger(agent.TriggerEvent{
			DeviceID: payload.DeviceID, Source: "threshold",
			Reading: &payload, PrevSoilMoisture: prevSoil,
		})
		return
	}
	if prevSoil != nil && math.Abs(*prevSoil-payload.SoilMoisture) > s.cfg.SoilMoistureDeltaPct {
		s.agent.Trigger(agent.TriggerEvent{
			DeviceID: payload.DeviceID, Source: "delta",
			Reading: &payload, PrevSoilMoisture: prevSoil,
		})
	}
}
