-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE party_movies ADD COLUMN watch_date DATE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE party_movies DROP COLUMN watch_date;
-- +goose StatementEnd
