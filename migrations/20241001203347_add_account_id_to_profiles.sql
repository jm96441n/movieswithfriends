-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
alter TABLE profiles ADD column id_account int;
CREATE INDEX idx_profiles_id_account ON profiles(id_account);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP INDEX idx_profiles_id_account;
alter TABLE profiles DROP column id_account;
-- +goose StatementEnd
