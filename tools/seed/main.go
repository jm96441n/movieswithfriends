package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"net/url"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	iamstore "github.com/jm96441n/movieswithfriends/identityaccess/store"
	partymgmtstore "github.com/jm96441n/movieswithfriends/partymgmt/store"
)

// TODO: add subcommands to nuke the DB and make seeding configurable (maybe with a file)
func main() {
	var shouldDrop bool
	flag.BoolVar(&shouldDrop, "drop", false, "drop the database before seeding")
	flag.Parse()

	connPool, err := setupDB()
	if err != nil {
		log.Fatalf("failed to setup db: %v", err)
	}

	if shouldDrop {
		err = dropDB(context.Background(), connPool)
		if err != nil {
			log.Fatalf("failed to drop db: %v", err)
		}
	}

	accountOneProfileInfo, err := seedAccountAndProfile(connPool, "test1@test.com", "1Password")
	if err != nil {
		log.Fatalf("failed to seed user1: %v", err)
	}

	accountTwoProfileInfo, err := seedAccountAndProfile(connPool, "test2@test.com", "2Password")
	if err != nil {
		log.Fatalf("failed to seed user2: %v", err)
	}

	// profile not in any party
	_, err = seedAccountAndProfile(connPool, "test3@test.com", "3Password")
	if err != nil {
		log.Fatalf("failed to seed user2: %v", err)
	}

	party1ID, err := seedPartyWithOwner(connPool, "party1", accountOneProfileInfo.ProfileID)
	if err != nil {
		log.Fatalf("failed to seed party1: %v", err)
	}

	party2ID, err := seedPartyWithOwner(connPool, "party2", accountOneProfileInfo.ProfileID)
	if err != nil {
		log.Fatalf("failed to seed party2: %v", err)
	}

	party3ID, err := seedPartyWithOwner(connPool, "party3", accountOneProfileInfo.ProfileID)
	if err != nil {
		log.Fatalf("failed to seed party3: %v", err)
	}

	for _, partyID := range []int{party1ID, party2ID, party3ID} {
		err = seedPartyMember(connPool, accountTwoProfileInfo.ProfileID, partyID)
		if err != nil {
			if errors.Is(err, partymgmtstore.ErrMemberPartyCombinationNotUnique) {
				slog.Info("profile in party alread", slog.Int("party_id", partyID), slog.Int("profile_id", accountTwoProfileInfo.ProfileID))
				continue
			}
			log.Fatalf("failed to seed party member with party ID %d: %v", partyID, err)
		}
	}
}

func seedAccountAndProfile(db *pgxpool.Pool, email, pw string) (iamstore.CreateProfileResult, error) {
	repo := iamstore.NewProfileRepository(db)
	hashed, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return iamstore.CreateProfileResult{}, err
	}

	profile, err := repo.GetProfileByEmail(context.Background(), email)

	if errors.Is(err, iamstore.ErrNoRecord) {
		result, err := repo.CreateProfile(context.Background(), email, "User1", "LastName1", hashed)
		if err != nil {
			return iamstore.CreateProfileResult{}, err
		}
		return result, nil
	}

	if err != nil {
		// profile doesn't exist so create it
		return iamstore.CreateProfileResult{}, err
	}

	return iamstore.CreateProfileResult{
		ProfileID: profile.ID,
		AccountID: profile.AccountID,
	}, nil
}

func seedPartyWithOwner(db *pgxpool.Pool, partyName string, ownerID int) (int, error) {
	repo := partymgmtstore.NewPartyRepository(db)
	return repo.CreateParty(context.Background(), ownerID, partyName, generateRandomString())
}

func seedPartyMember(db *pgxpool.Pool, profileID, partyID int) error {
	repo := partymgmtstore.NewPartyRepository(db)
	return repo.CreatePartyMember(context.Background(), profileID, partyID)
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// generate a random 6 character string
func generateRandomString() string {
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

type DBCreds struct {
	Username string
	Password string
}

func setupDB() (*pgxpool.Pool, error) {
	dbUser := os.Getenv("DB_MIGRATION_USER")
	if dbUser == "" {
		return nil, ErrMissingDBUsername
	}

	dbPassword := os.Getenv("DB_MIGRATION_PASSWORD")
	if dbPassword == "" {
		return nil, ErrMissingDBPassword
	}

	creds, err := newDBCreds(dbUser, dbPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to create new db creds: %w", err)
	}

	connPool, err := createConnPool("0.0.0.0", os.Getenv("DB_DATABASE_NAME"), creds)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}
	return connPool, nil
}

func newDBCreds(username, pw string) (DBCreds, error) {
	if username == "" {
		return DBCreds{}, ErrMissingDBUsername
	}

	if pw == "" {
		return DBCreds{}, ErrMissingDBPassword
	}
	return DBCreds{
		Username: username,
		Password: url.QueryEscape(pw),
	}, nil
}

func createConnPool(host string, dbname string, creds DBCreds) (*pgxpool.Pool, error) {
	if host == "" {
		return nil, ErrMissingDBHost
	}

	if dbname == "" {
		return nil, ErrMissingDBDatabaseName
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	connString := fmt.Sprintf("postgres://%s:%s@%s/%s", creds.Username, creds.Password, host, dbname)
	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	db, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	err = db.Ping(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func dropDB(ctx context.Context, conn *pgxpool.Pool) error {
	_, err := conn.Exec(ctx, `DO $$ 
DECLARE 
    r RECORD;
BEGIN
    -- Disable foreign key checks temporarily
    SET CONSTRAINTS ALL DEFERRED;
    
    FOR r IN (SELECT tablename 
              FROM pg_tables 
              WHERE schemaname = 'public' 
              AND tablename != 'goose_db_version') LOOP
        EXECUTE 'TRUNCATE TABLE ' || quote_ident(r.tablename) || ' CASCADE';
    END LOOP;
END $$;`)
	if err != nil {
		return err
	}

	return nil
}

var (
	ErrMissingDBUsername     = errors.New("DB_USERNAME env var is missing")
	ErrMissingDBPassword     = errors.New("DB_PASSWORD env var is missing")
	ErrMissingDBHost         = errors.New("DB_HOST env var is missing")
	ErrMissingDBDatabaseName = errors.New("DB_DATABASE_NAME env var is missing")
)
