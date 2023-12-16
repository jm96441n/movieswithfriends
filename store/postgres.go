package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Creds struct {
	Username string
	Password string
}

type PGStore struct {
	db *pgxpool.Pool
}

var (
	ErrMissingDBUsername     = errors.New("DB_USERNAME env var is missing")
	ErrMissingDBPassword     = errors.New("DB_PASSWORD env var is missing")
	ErrMissingDBHost         = errors.New("DB_HOST env var is missing")
	ErrMissingDBDatabaseName = errors.New("DB_DATABASE_NAME env var is missing")
)

func NewCreds(username, pw string) (Creds, error) {
	if username == "" {
		return Creds{}, ErrMissingDBUsername
	}

	if pw == "" {
		return Creds{}, ErrMissingDBPassword
	}
	return Creds{
		Username: username,
		Password: pw,
	}, nil
}

func NewPostgesStore(creds Creds, host, dbname string) (*PGStore, error) {
	if host == "" {
		return nil, ErrMissingDBHost
	}

	if dbname == "" {
		return nil, ErrMissingDBDatabaseName
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()

	connString := fmt.Sprintf("postgres://%s:%s@%s/%s", creds.Username, creds.Password, host, dbname)
	db, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}

	db.Ping(ctx)

	return &PGStore{db: db}, nil
}

func (pg *PGStore) Ping(ctx context.Context) error {
	return pg.db.Ping(ctx)
}

func (pg *PGStore) Close() {
	pg.db.Close()
}
