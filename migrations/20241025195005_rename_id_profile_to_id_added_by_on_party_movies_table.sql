-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE party_movies RENAME COLUMN id_profile TO id_added_by;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE party_movies RENAME COLUMN id_added_by TO id_profile;
-- +goose StatementEnd
