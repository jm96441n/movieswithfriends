-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

CREATE TABLE invitations (
    id_invitation INT GENERATED ALWAYS AS IDENTITY,
    id_profile INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (clock_timestamp() AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP,
    PRIMARY KEY(id_invitation),
    FOREIGN KEY (id_profile) REFERENCES profiles(id_profile)
);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

DROP TABLE IF EXISTS invitations;
