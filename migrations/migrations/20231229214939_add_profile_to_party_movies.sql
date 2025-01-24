-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
alter table party_movies add column id_profile int;
alter table party_movies add CONSTRAINT fk_party_movies_profile FOREIGN KEY(id_profile) REFERENCES profiles(id_profile);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
alter table party_movies drop column id_profile;
-- +goose StatementEnd
