package store

import (
	"context"

	"agriagent/backend/internal/models"
)

// Store abstracts persistence so handlers/agent can run against either the
// in-memory MockStore (no DB required, used for early frontend integration)
// or PostgresStore (real Supabase-backed storage).
type Store interface {
	ListDevices(ctx context.Context) ([]models.Device, error)
	TouchDevice(ctx context.Context, deviceID string) error // marks device active + last_seen=now

	LatestReadings(ctx context.Context) ([]models.SensorReading, error) // most recent reading per device
	ReadingsRange(ctx context.Context, rangeKey string) ([]models.SensorReading, error)
	LatestReadingForDevice(ctx context.Context, deviceID string) (*models.SensorReading, bool, error)
	InsertReading(ctx context.Context, p models.SensorPayload) (models.SensorReading, error)

	RecentImages(ctx context.Context, limit int) ([]models.LeafImage, error)

	ListDecisions(ctx context.Context) ([]models.AgentDecision, error)
	GetDecision(ctx context.Context, id int64) (models.AgentDecision, bool, error)
	InsertDecision(ctx context.Context, d models.AgentDecision) (models.AgentDecision, error)
	ConfirmDecision(ctx context.Context, id int64, confirmedBy string, approve bool) (models.AgentDecision, error)
}
