-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

create table accounts (
    account_id INT GENERATED ALWAYS AS IDENTITY,
    login VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMP,
    PRIMARY KEY(account_id)
);

create unique index accounts_login_uniq_idx ON accounts(LOWER(login));

create index accounts_login_idx ON accounts(LOWER(login));

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
drop index if exists accounts_login_idx;

drop index if exists accounts_login_uniq_idx;

drop table if exists accounts;
