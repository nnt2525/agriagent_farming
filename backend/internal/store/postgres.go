package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"agriagent/backend/internal/models"
)

// PostgresStore implements Store against the schema in migrations/0001_init.sql.
type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(ctx context.Context, databaseURL string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &PostgresStore{pool: pool}, nil
}

func (s *PostgresStore) Close() {
	s.pool.Close()
}

func (s *PostgresStore) ListDevices(ctx context.Context) ([]models.Device, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, type, zone, status, last_seen FROM devices ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Device
	for rows.Next() {
		var d models.Device
		if err := rows.Scan(&d.ID, &d.Name, &d.Type, &d.Zone, &d.Status, &d.LastSeen); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *PostgresStore) TouchDevice(ctx context.Context, deviceID string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE devices SET status = 'active', last_seen = now() WHERE id = $1`, deviceID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		_, err = s.pool.Exec(ctx, `
			INSERT INTO devices (id, name, type, zone, status, last_seen)
			VALUES ($1, $1, 'soil_sensor', 'unassigned', 'active', now())
			ON CONFLICT (id) DO UPDATE SET status = 'active', last_seen = now()`, deviceID)
	}
	return err
}

func (s *PostgresStore) LatestReadings(ctx context.Context) ([]models.SensorReading, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT ON (device_id) id, device_id, soil_moisture, temperature, humidity, created_at
		FROM sensor_readings
		ORDER BY device_id, created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanReadings(rows)
}

func (s *PostgresStore) ReadingsRange(ctx context.Context, rangeKey string) ([]models.SensorReading, error) {
	var interval string
	switch rangeKey {
	case "month":
		interval = "1 month"
	case "year":
		interval = "1 year"
	default:
		interval = "1 day"
	}
	rows, err := s.pool.Query(ctx, fmt.Sprintf(`
		SELECT id, device_id, soil_moisture, temperature, humidity, created_at
		FROM sensor_readings
		WHERE created_at > now() - interval '%s'
		ORDER BY created_at ASC`, interval))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanReadings(rows)
}

func (s *PostgresStore) LatestReadingForDevice(ctx context.Context, deviceID string) (*models.SensorReading, bool, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, device_id, soil_moisture, temperature, humidity, created_at
		FROM sensor_readings WHERE device_id = $1
		ORDER BY created_at DESC LIMIT 1`, deviceID)
	var r models.SensorReading
	err := row.Scan(&r.ID, &r.DeviceID, &r.SoilMoisture, &r.Temperature, &r.Humidity, &r.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return &r, true, nil
}

func (s *PostgresStore) InsertReading(ctx context.Context, p models.SensorPayload) (models.SensorReading, error) {
	ts := p.Timestamp
	if ts.IsZero() {
		ts = time.Now()
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO sensor_readings (device_id, soil_moisture, temperature, humidity, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, device_id, soil_moisture, temperature, humidity, created_at`,
		p.DeviceID, p.SoilMoisture, p.Temperature, p.Humidity, ts)
	var r models.SensorReading
	err := row.Scan(&r.ID, &r.DeviceID, &r.SoilMoisture, &r.Temperature, &r.Humidity, &r.CreatedAt)
	return r, err
}

func (s *PostgresStore) RecentImages(ctx context.Context, limit int) ([]models.LeafImage, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, device_id, image_url, COALESCE(cv_result, ''), COALESCE(confidence, 0), created_at
		FROM leaf_images ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.LeafImage
	for rows.Next() {
		var img models.LeafImage
		if err := rows.Scan(&img.ID, &img.DeviceID, &img.ImageURL, &img.CVResult, &img.Confidence, &img.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, img)
	}
	return out, rows.Err()
}

func (s *PostgresStore) ListDecisions(ctx context.Context) ([]models.AgentDecision, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, COALESCE(device_id::text, ''), action, reason, COALESCE(confidence, 0),
		       need_human_confirm, confirmed_by, confirmed_at, status, trigger_source, created_at
		FROM agent_decisions ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDecisions(rows)
}

func (s *PostgresStore) GetDecision(ctx context.Context, id int64) (models.AgentDecision, bool, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, COALESCE(device_id::text, ''), action, reason, COALESCE(confidence, 0),
		       need_human_confirm, confirmed_by, confirmed_at, status, trigger_source, created_at
		FROM agent_decisions WHERE id = $1`, id)
	d, err := scanDecisionRow(row)
	if err == pgx.ErrNoRows {
		return models.AgentDecision{}, false, nil
	}
	if err != nil {
		return models.AgentDecision{}, false, err
	}
	return d, true, nil
}

func (s *PostgresStore) InsertDecision(ctx context.Context, d models.AgentDecision) (models.AgentDecision, error) {
	status := d.Status
	if status == "" {
		if d.NeedHumanConfirm {
			status = "pending"
		} else {
			status = "auto_executed"
		}
	}
	var deviceID interface{}
	if d.DeviceID != "" {
		deviceID = d.DeviceID
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO agent_decisions (device_id, action, reason, confidence, need_human_confirm, status, trigger_source)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, COALESCE(device_id::text, ''), action, reason, COALESCE(confidence, 0),
		          need_human_confirm, confirmed_by, confirmed_at, status, trigger_source, created_at`,
		deviceID, d.Action, d.Reason, d.Confidence, d.NeedHumanConfirm, status, d.TriggerSource)
	return scanDecisionRow(row)
}

func (s *PostgresStore) ConfirmDecision(ctx context.Context, id int64, confirmedBy string, approve bool) (models.AgentDecision, error) {
	status := "rejected"
	if approve {
		status = "confirmed"
	}
	row := s.pool.QueryRow(ctx, `
		UPDATE agent_decisions
		SET confirmed_by = $2, confirmed_at = now(), status = $3
		WHERE id = $1
		RETURNING id, COALESCE(device_id::text, ''), action, reason, COALESCE(confidence, 0),
		          need_human_confirm, confirmed_by, confirmed_at, status, trigger_source, created_at`,
		id, confirmedBy, status)
	return scanDecisionRow(row)
}

func scanReadings(rows pgx.Rows) ([]models.SensorReading, error) {
	var out []models.SensorReading
	for rows.Next() {
		var r models.SensorReading
		if err := rows.Scan(&r.ID, &r.DeviceID, &r.SoilMoisture, &r.Temperature, &r.Humidity, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func scanDecisions(rows pgx.Rows) ([]models.AgentDecision, error) {
	var out []models.AgentDecision
	for rows.Next() {
		d, err := scanDecisionRowFromRows(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func scanDecisionRowFromRows(rows pgx.Rows) (models.AgentDecision, error) {
	var d models.AgentDecision
	err := rows.Scan(&d.ID, &d.DeviceID, &d.Action, &d.Reason, &d.Confidence,
		&d.NeedHumanConfirm, &d.ConfirmedBy, &d.ConfirmedAt, &d.Status, &d.TriggerSource, &d.CreatedAt)
	return d, err
}

func scanDecisionRow(row pgx.Row) (models.AgentDecision, error) {
	var d models.AgentDecision
	err := row.Scan(&d.ID, &d.DeviceID, &d.Action, &d.Reason, &d.Confidence,
		&d.NeedHumanConfirm, &d.ConfirmedBy, &d.ConfirmedAt, &d.Status, &d.TriggerSource, &d.CreatedAt)
	return d, err
}
