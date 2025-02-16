-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE parties ADD COLUMN id_owner int;
ALTER TABLE parties ADD CONSTRAINT fk_owner FOREIGN KEY (id_owner) REFERENCES profiles (id_profile);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE parties DROP CONSTRAINT fk_owner;
ALTER TABLE parties DROP COLUMN id_owner;
-- +goose StatementEnd
