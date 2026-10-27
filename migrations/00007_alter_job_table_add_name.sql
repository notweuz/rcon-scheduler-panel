-- +goose Up
ALTER TABLE jobs
    ADD COLUMN name VARCHAR(50) NOT NULL;

-- +goose Down
ALTER TABLE jobs DROP COLUMN name;