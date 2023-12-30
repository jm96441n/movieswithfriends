-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

create table profile_parties (
    id_profile_party INT GENERATED ALWAYS AS IDENTITY,
    id_profile INT NOT NULL,
    id_party INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMP,
    PRIMARY KEY(id_profile_party),
    CONSTRAINT fk_profile_parties_profiles FOREIGN KEY(id_profile) REFERENCES profiles(id_profile),
    CONSTRAINT fk_profile_parties_parties FOREIGN KEY(id_party) REFERENCES parties(id_party)
);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

drop table if exists profile_parties;
