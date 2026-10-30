-- +goose Up
ALTER TABLE protocols 
DROP COLUMN default_port;

-- +goose Down
ALTER TABLE protocols 
ADD COLUMN default_port INTEGER;