-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE parties ADD CONSTRAINT unique_parties_short_id UNIQUE(short_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE parties DROP CONSTRAINT unique_parties_short_id UNIQUE(short_id);
-- +goose StatementEnd
