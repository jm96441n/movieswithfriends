-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

create table party_members (
    id_profile_party INT GENERATED ALWAYS AS IDENTITY,
    id_member INT NOT NULL,
    id_party INT NOT NULL,
    owner BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMP,
    PRIMARY KEY(id_profile_party),
    CONSTRAINT fk_party_members_members FOREIGN KEY(id_member) REFERENCES profiles(id_profile),
    CONSTRAINT fk_party_members_parties FOREIGN KEY(id_party) REFERENCES parties(id_party)
);

alter table party_members add CONSTRAINT unique_member_per_party UNIQUE(id_party, id_member);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
alter table party_members drop CONSTRAINT unique_member_per_party;
drop table if exists party_members;
