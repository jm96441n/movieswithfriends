-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

create table profiles (
    profile_id INT GENERATED ALWAYS AS IDENTITY,
    name VARCHAR(255) NOT NULL;
    account_id INT,
    group_id INT,
    created_at TIMESTAMP NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMP,
    PRIMARY KEY(profile_id),
    CONSTRAINT fk_account FOREIGN KEY(account_id) REFERENCES tableName(accounts)
    CONSTRAINT fk_group FOREIGN KEY(group_id) REFERENCES tableName(groups)
);

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd


drop table if exists profiles;
