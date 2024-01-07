-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TYPE watch_status AS ENUM ('unwatched', 'selected', 'watched');
ALTER TABLE party_movies ADD COLUMN watch_status watch_status NOT NULL DEFAULT 'unwatched';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE party_movies DROP COLUMN watch_status;
DROP TYPE watch_status;
-- +goose StatementEnd
