-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE party_members DROP COLUMN owner;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE party_members ADD COLUMN owner bool;
-- +goose StatementEnd
