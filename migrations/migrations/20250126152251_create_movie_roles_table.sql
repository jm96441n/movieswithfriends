-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE movie_roles (
    id_movie_role INT GENERATED ALWAYS AS IDENTITY,
    id_movie INT NOT NULL,
    id_people INT NOT NULL,
    character_name VARCHAR(100) NOT NULL,
    job VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMP,
    PRIMARY KEY(id_movie_role),
    FOREIGN KEY (id_movie) REFERENCES movies(id_movie),
    FOREIGN KEY (id_people) REFERENCES people(id_people)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
