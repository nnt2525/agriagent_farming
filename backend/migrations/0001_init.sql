-- AgriAgent initial schema
-- Run against PostgreSQL (Supabase compatible)

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS devices (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    type        TEXT NOT NULL,              -- e.g. 'soil_sensor', 'camera', 'relay'
    zone        TEXT NOT NULL,               -- farm zone / plot identifier
    status      TEXT NOT NULL DEFAULT 'inactive', -- 'active' | 'inactive'
    last_seen   TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sensor_readings (
    id             BIGSERIAL PRIMARY KEY,
    device_id      UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    soil_moisture  DOUBLE PRECISION NOT NULL,
    temperature    DOUBLE PRECISION NOT NULL,
    humidity       DOUBLE PRECISION NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_sensor_readings_device_created
    ON sensor_readings (device_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_sensor_readings_created
    ON sensor_readings (created_at DESC);

CREATE TABLE IF NOT EXISTS leaf_images (
    id          BIGSERIAL PRIMARY KEY,
    device_id   UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    image_url   TEXT NOT NULL,
    cv_result   TEXT,                        -- e.g. 'healthy', 'blight', 'yellowing'
    confidence  DOUBLE PRECISION,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_leaf_images_created
    ON leaf_images (created_at DESC);

CREATE TABLE IF NOT EXISTS agent_decisions (
    id                  BIGSERIAL PRIMARY KEY,
    device_id           UUID REFERENCES devices(id) ON DELETE SET NULL,
    action              TEXT NOT NULL,        -- e.g. 'water_on', 'water_off', 'no_action', 'alert'
    reason              TEXT NOT NULL,
    confidence          DOUBLE PRECISION,
    need_human_confirm  BOOLEAN NOT NULL DEFAULT false,
    confirmed_by        TEXT,                 -- admin identifier, null until confirmed
    confirmed_at        TIMESTAMPTZ,
    status              TEXT NOT NULL DEFAULT 'pending', -- 'pending' | 'confirmed' | 'rejected' | 'auto_executed'
    trigger_source      TEXT NOT NULL DEFAULT 'event',   -- 'threshold' | 'delta' | 'cv' | 'schedule' | 'manual'
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_agent_decisions_created
    ON agent_decisions (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_agent_decisions_pending
    ON agent_decisions (need_human_confirm, status);
