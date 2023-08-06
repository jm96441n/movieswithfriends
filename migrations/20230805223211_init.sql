-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

create table accounts (
    id bigserial primary key,
    login varchar(255) not null,
    password_hash varchar(255) not null,
    created_at timestamptz not null default clock_timestamp(),
    updated_at timestamptz
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
