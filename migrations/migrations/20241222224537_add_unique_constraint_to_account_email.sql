-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE accounts ADD CONSTRAINT unique_email_addresses UNIQUE(email);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE accounts DROP CONSTRAINT unique_email_addresses UNIQUE(email);
-- +goose StatementEnd
