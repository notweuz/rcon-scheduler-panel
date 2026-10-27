-- +goose Up
ALTER TABLE jobs 
    ADD COLUMN steps JSONB NOT NULL DEFAULT '[]'::jsonb
        CHECK (jsonb_typeof(steps) = 'array');

-- +goose Down
ALTER TABLE jobs DROP COLUMN steps;

-- Initially I wanted to make jobs have it's own table,
-- but they're definitely never gonna be reused or shared between different jobs;
-- probably will stay like this forever unless I want to rework how tasks are stored