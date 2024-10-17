-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE parties ADD column short_id text NOT NULL default substring(md5(random()::text), 0, 8);
CREATE INDEX idx_parties_short_id ON parties(short_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP INDEX idx_parties_short_id;
ALTER TABLE parties DROP column short_id;
-- +goose StatementEnd
