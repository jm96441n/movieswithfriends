-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER table party_movies ADD CONSTRAINT fk_party_movies_added_by FOREIGN KEY(id_added_by) REFERENCES profiles(id_profile),
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER table party_movies DROP CONSTRAINT fk_party_movies_added_by;
-- +goose StatementEnd
