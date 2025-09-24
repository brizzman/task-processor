-- +goose Up
-- +goose StatementBegin
CREATE TABLE failed_tasks (
    id UUID PRIMARY KEY,
    status task_status NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    attempts INTEGER NOT NULL,
    max_attempts INTEGER NOT NULL,
    error_message TEXT NULL,
    moved_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_failed_tasks_moved_at ON failed_tasks (moved_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS failed_tasks;
DROP INDEX IF EXISTS idx_failed_tasks_moved_at;
-- +goose StatementEnd
