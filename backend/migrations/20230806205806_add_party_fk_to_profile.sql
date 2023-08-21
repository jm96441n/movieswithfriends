-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

ALTER table profiles
ADD CONSTRAINT fk_party FOREIGN KEY(id_party) REFERENCES parties(id_party);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd


ALTER table profiles
DROP CONSTRAINT fk_party;
