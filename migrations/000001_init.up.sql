-- migrations/000001_init.up.sql

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users
CREATE TABLE users (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email                TEXT NOT NULL UNIQUE,
    password_hash        TEXT NOT NULL,
    name                 TEXT NOT NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    pref_nail_style      TEXT DEFAULT 'top_mounted',
    pref_nail_diameter_mm NUMERIC(4,2) DEFAULT 1.5,
    pref_units           TEXT DEFAULT 'metric',
    pref_auto_save       BOOLEAN DEFAULT TRUE,
    pref_haptic          BOOLEAN DEFAULT FALSE
);

-- Refresh tokens
CREATE TABLE refresh_tokens (
    token      TEXT PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);

-- Projects
CREATE TABLE projects (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title            TEXT NOT NULL,
    shape            TEXT NOT NULL DEFAULT 'circle',
    size_inches      NUMERIC(6,2) NOT NULL DEFAULT 12,
    nail_count       INTEGER NOT NULL DEFAULT 200,
    nail_style       TEXT NOT NULL DEFAULT 'top_mounted',
    nail_diameter_mm NUMERIC(4,2) NOT NULL DEFAULT 1.5,
    layer_mode       BOOLEAN NOT NULL DEFAULT FALSE,
    layer_count      INTEGER NOT NULL DEFAULT 1,
    image_remote_url TEXT NOT NULL DEFAULT '',
    string_plan_json TEXT NOT NULL DEFAULT '{}',
    status           TEXT NOT NULL DEFAULT 'pending',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_projects_user_id ON projects(user_id);

-- Progress
CREATE TABLE project_progress (
    project_id   UUID PRIMARY KEY REFERENCES projects(id) ON DELETE CASCADE,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    current_step INTEGER NOT NULL DEFAULT 0,
    total_steps  INTEGER NOT NULL DEFAULT 0,
    last_updated TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

-- Progress markers
CREATE TABLE progress_markers (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    step       INTEGER NOT NULL,
    label      TEXT NOT NULL DEFAULT '',
    note       TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_progress_markers_project_id ON progress_markers(project_id);
