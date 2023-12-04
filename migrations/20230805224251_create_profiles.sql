-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

create table profiles (
    id_profile INT GENERATED ALWAYS AS IDENTITY,
    name VARCHAR(255) NOT NULL,
    id_account INT,
    id_party INT,
    created_at TIMESTAMP NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMP,
    PRIMARY KEY(id_profile),
    CONSTRAINT fk_account FOREIGN KEY(id_account) REFERENCES accounts(id_account)
);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

drop table if exists profiles;
