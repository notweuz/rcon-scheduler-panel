-- +goose Up
CREATE INDEX idx_jobs_server_id ON jobs(server_id);

-- +goose Down
DROP INDEX IF EXISTS idx_jobs_server_id;