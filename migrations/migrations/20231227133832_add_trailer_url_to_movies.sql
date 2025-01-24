-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
alter table movies add column trailer_url varchar(255);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
alter table movies drop column trailer_url;
-- +goose StatementEnd
