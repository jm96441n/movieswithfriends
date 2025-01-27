-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE movies ADD COLUMN budget integer;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE movies DROP COLUMN budget;
-- +goose StatementEnd
