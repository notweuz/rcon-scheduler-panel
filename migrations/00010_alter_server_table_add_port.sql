-- +goose Up
ALTER TABLE servers
ADD COLUMN port INTEGER NOT NULL DEFAULT 25575;

-- +goose Down
ALTER TABLE servers
DROP COLUMN port;