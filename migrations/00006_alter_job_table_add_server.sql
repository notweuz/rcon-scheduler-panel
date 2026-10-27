-- +goose Up
ALTER TABLE jobs
    ADD COLUMN server_id BIGINT NOT NULL REFERENCES servers(id) ON DELETE CASCADE;

-- +goose Down
ALTER TABLE jobs DROP COLUMN server_id;