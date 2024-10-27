-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER table movies 
  ADD COLUMN runtime INT, 
  ADD COLUMN rating FLOAT,
  ADD COLUMN genres TEXT[];

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER table movies 
  DROP COLUMN runtime INT, 
  DROP COLUMN rating FLOAT,
  DROP COLUMN genres TEXT[];
-- +goose StatementEnd
