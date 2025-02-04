-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
ALTER TABLE invitations ADD COLUMN email VARCHAR(255) NOT NULL;
ALTER TABLE invitations ALTER COLUMN id_profile DROP NOT NULL;
ALTER TABLE invitations ADD COLUMN id_party INT NOT NULL;
ALTER TABLE invitations ADD CONSTRAINT fk_invitations_party FOREIGN KEY(id_party) REFERENCES parties(id_party);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
ALTER TABLE invitations DROP COLUMN email;
ALTER TABLE invitations DROP COLUMN id_party;
