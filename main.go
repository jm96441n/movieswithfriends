package main

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	localhttp "github.com/jm96441n/movieswithfriends/http"
	"golang.org/x/exp/slog"
)

func main() {
	//	urlExample := "postgres://myuser:password@localhost:5432/database_name"
	//	conn, err := pgx.Connect(context.Background(), urlExample)
	//	if err != nil {
	//		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
	//		os.Exit(1)
	//	}
	//	defer conn.Close(context.Background())
	//
	//	err = conn.QueryRow(context.Background(), "select * from accounts;")
	//	if err != nil {
	//		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
	//		os.Exit(1)
	//	}
	//
	tmpls := localhttp.BuildTemplates()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	router := mux.NewRouter()

	localhttp.SetupWebServer(logger, router, tmpls)
	logger.Info("Listening on :8080...")

	if err := http.ListenAndServe(":8080", router); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
