-- +goose Up
-- +goose StatementBegin
CREATE TYPE task_status AS ENUM (
    'NEW',
    'PROCESSING', 
    'PROCESSED',
    'FAILED'
);

CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status task_status NOT NULL DEFAULT 'NEW'::task_status,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    attempts INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 3,
    error_message TEXT NULL
);

CREATE INDEX idx_tasks_ready ON tasks (status, created_at) WHERE attempts < max_attempts;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tasks;
DROP TYPE IF EXISTS task_status;
DROP INDEX IF EXISTS idx_tasks_ready;
-- +goose StatementEnd
