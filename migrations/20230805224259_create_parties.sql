-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

create table parties (
    id_party INT GENERATED ALWAYS AS IDENTITY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMP,
    PRIMARY KEY(id_party)
);

create index party_name_idx ON parties(LOWER(name));
create unique index party_name_uniq_idx ON parties(LOWER(name));

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

drop index if exists party_name__idx;
drop index if exists party_name_uniq_idx;
drop table if exists parties;
