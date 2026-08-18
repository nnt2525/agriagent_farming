package store

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"agriagent/backend/internal/models"
)

// MockStore is an in-memory Store seeded with plausible sample data so the
// frontend can be built against real-shaped responses before Postgres/MQTT
// are connected. Safe for concurrent use.
type MockStore struct {
	mu sync.Mutex

	devices   map[string]*models.Device
	readings  []models.SensorReading
	images    []models.LeafImage
	decisions []models.AgentDecision

	nextReadingID  int64
	nextDecisionID int64
}

func NewMockStore() *MockStore {
	s := &MockStore{
		devices: map[string]*models.Device{},
	}
	s.seed()
	return s
}

func (s *MockStore) seed() {
	now := time.Now()

	devSpecs := []struct {
		id, name, typ, zone, status string
		lastSeenAgo                 time.Duration
	}{
		{"esp32-soil-01", "Soil Sensor Zone A", "soil_sensor", "Zone A", "active", 2 * time.Minute},
		{"esp32-soil-02", "Soil Sensor Zone B", "soil_sensor", "Zone B", "active", 4 * time.Minute},
		{"esp32-cam-01", "Leaf Camera Zone A", "camera", "Zone A", "active", 90 * time.Minute},
		{"esp32-relay-01", "Water Relay Zone A", "relay", "Zone A", "active", 5 * time.Minute},
		{"esp32-soil-03", "Soil Sensor Zone C", "soil_sensor", "Zone C", "inactive", 26 * time.Hour},
	}
	for _, d := range devSpecs {
		ls := now.Add(-d.lastSeenAgo)
		s.devices[d.id] = &models.Device{
			ID: d.id, Name: d.name, Type: d.typ, Zone: d.zone, Status: d.status, LastSeen: &ls,
		}
	}

	// Seed ~30 days of readings, every 30 min, for the two active soil sensors.
	soilDevices := []string{"esp32-soil-01", "esp32-soil-02"}
	start := now.Add(-30 * 24 * time.Hour)
	id := int64(1)
	for t := start; t.Before(now); t = t.Add(30 * time.Minute) {
		for i, devID := range soilDevices {
			base := 45.0 + 10*math.Sin(float64(t.Unix())/86400*2*math.Pi) + float64(i)*3
			s.readings = append(s.readings, models.SensorReading{
				ID:           id,
				DeviceID:     devID,
				SoilMoisture: clamp(base+rand.Float64()*6-3, 5, 95),
				Temperature:  clamp(27+5*math.Sin(float64(t.Unix())/86400*2*math.Pi)+rand.Float64()*2-1, 15, 40),
				Humidity:     clamp(60+15*math.Cos(float64(t.Unix())/86400*2*math.Pi)+rand.Float64()*4-2, 20, 95),
				CreatedAt:    t,
			})
			id++
		}
	}
	s.nextReadingID = id

	// Seed leaf images every 3h for the last 2 days.
	cvResults := []string{"healthy", "healthy", "healthy", "yellowing", "healthy"}
	imgStart := now.Add(-48 * time.Hour)
	imgID := int64(1)
	for t := imgStart; t.Before(now); t = t.Add(3 * time.Hour) {
		res := cvResults[rand.Intn(len(cvResults))]
		s.images = append(s.images, models.LeafImage{
			ID:         imgID,
			DeviceID:   "esp32-cam-01",
			ImageURL:   fmt.Sprintf("https://picsum.photos/seed/leaf-%d/480/360", imgID),
			CVResult:   res,
			Confidence: clamp(0.7+rand.Float64()*0.29, 0, 1),
			CreatedAt:  t,
		})
		imgID++
	}

	// Seed a few agent decisions across states.
	seedDecisions := []models.AgentDecision{
		{
			DeviceID: "esp32-relay-01", Action: "water_on",
			Reason: "Soil moisture 22% below 30% threshold in Zone A; forecast shows no rain in next 6h.",
			Confidence: 0.93, NeedHumanConfirm: false, Status: "auto_executed",
			TriggerSource: "threshold", CreatedAt: now.Add(-6 * time.Hour),
		},
		{
			DeviceID: "esp32-cam-01", Action: "alert",
			Reason: "CV detected possible early blight on leaf image (yellowing pattern), confidence borderline.",
			Confidence: 0.58, NeedHumanConfirm: true, Status: "pending",
			TriggerSource: "cv", CreatedAt: now.Add(-2 * time.Hour),
		},
		{
			DeviceID: "esp32-relay-01", Action: "no_action",
			Reason: "Soil moisture stable at 52%, within healthy range for tomato at this growth stage.",
			Confidence: 0.88, NeedHumanConfirm: false, Status: "auto_executed",
			TriggerSource: "schedule", CreatedAt: now.Add(-3 * time.Hour),
		},
		{
			DeviceID: "esp32-relay-01", Action: "water_on",
			Reason: "Sudden 14% drop in soil moisture within 10 minutes; possible sensor fault or drainage issue, requesting confirmation before irrigating.",
			Confidence: 0.61, NeedHumanConfirm: true, Status: "pending",
			TriggerSource: "delta", CreatedAt: now.Add(-30 * time.Minute),
		},
	}
	for i := range seedDecisions {
		seedDecisions[i].ID = int64(i + 1)
	}
	s.decisions = seedDecisions
	s.nextDecisionID = int64(len(seedDecisions) + 1)
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func (s *MockStore) ListDevices(ctx context.Context) ([]models.Device, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]models.Device, 0, len(s.devices))
	for _, d := range s.devices {
		out = append(out, *d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (s *MockStore) TouchDevice(ctx context.Context, deviceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	if d, ok := s.devices[deviceID]; ok {
		d.LastSeen = &now
		d.Status = "active"
		return nil
	}
	s.devices[deviceID] = &models.Device{
		ID: deviceID, Name: deviceID, Type: "soil_sensor", Zone: "unassigned",
		Status: "active", LastSeen: &now,
	}
	return nil
}

func (s *MockStore) LatestReadings(ctx context.Context) ([]models.SensorReading, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	latest := map[string]models.SensorReading{}
	for _, r := range s.readings {
		if cur, ok := latest[r.DeviceID]; !ok || r.CreatedAt.After(cur.CreatedAt) {
			latest[r.DeviceID] = r
		}
	}
	out := make([]models.SensorReading, 0, len(latest))
	for _, r := range latest {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].DeviceID < out[j].DeviceID })
	return out, nil
}

func (s *MockStore) ReadingsRange(ctx context.Context, rangeKey string) ([]models.SensorReading, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	var since time.Time
	switch rangeKey {
	case "month":
		since = now.AddDate(0, -1, 0)
	case "year":
		since = now.AddDate(-1, 0, 0)
	default: // "day"
		since = now.AddDate(0, 0, -1)
	}
	out := make([]models.SensorReading, 0)
	for _, r := range s.readings {
		if r.CreatedAt.After(since) {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

func (s *MockStore) LatestReadingForDevice(ctx context.Context, deviceID string) (*models.SensorReading, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var found *models.SensorReading
	for i := range s.readings {
		r := s.readings[i]
		if r.DeviceID != deviceID {
			continue
		}
		if found == nil || r.CreatedAt.After(found.CreatedAt) {
			rc := r
			found = &rc
		}
	}
	return found, found != nil, nil
}

func (s *MockStore) InsertReading(ctx context.Context, p models.SensorPayload) (models.SensorReading, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ts := p.Timestamp
	if ts.IsZero() {
		ts = time.Now()
	}
	r := models.SensorReading{
		ID: s.nextReadingID, DeviceID: p.DeviceID,
		SoilMoisture: p.SoilMoisture, Temperature: p.Temperature, Humidity: p.Humidity,
		CreatedAt: ts,
	}
	s.nextReadingID++
	s.readings = append(s.readings, r)
	return r, nil
}

func (s *MockStore) RecentImages(ctx context.Context, limit int) ([]models.LeafImage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sorted := make([]models.LeafImage, len(s.images))
	copy(sorted, s.images)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].CreatedAt.After(sorted[j].CreatedAt) })
	if limit > 0 && len(sorted) > limit {
		sorted = sorted[:limit]
	}
	return sorted, nil
}

func (s *MockStore) ListDecisions(ctx context.Context) ([]models.AgentDecision, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sorted := make([]models.AgentDecision, len(s.decisions))
	copy(sorted, s.decisions)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].CreatedAt.After(sorted[j].CreatedAt) })
	return sorted, nil
}

func (s *MockStore) GetDecision(ctx context.Context, id int64) (models.AgentDecision, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, d := range s.decisions {
		if d.ID == id {
			return d, true, nil
		}
	}
	return models.AgentDecision{}, false, nil
}

func (s *MockStore) InsertDecision(ctx context.Context, d models.AgentDecision) (models.AgentDecision, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	d.ID = s.nextDecisionID
	s.nextDecisionID++
	if d.CreatedAt.IsZero() {
		d.CreatedAt = time.Now()
	}
	if d.Status == "" {
		if d.NeedHumanConfirm {
			d.Status = "pending"
		} else {
			d.Status = "auto_executed"
		}
	}
	s.decisions = append(s.decisions, d)
	return d, nil
}

func (s *MockStore) ConfirmDecision(ctx context.Context, id int64, confirmedBy string, approve bool) (models.AgentDecision, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.decisions {
		if s.decisions[i].ID != id {
			continue
		}
		now := time.Now()
		s.decisions[i].ConfirmedBy = &confirmedBy
		s.decisions[i].ConfirmedAt = &now
		if approve {
			s.decisions[i].Status = "confirmed"
		} else {
			s.decisions[i].Status = "rejected"
		}
		return s.decisions[i], nil
	}
	return models.AgentDecision{}, fmt.Errorf("decision %d not found", id)
}
