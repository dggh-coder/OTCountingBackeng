package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"otcountingbackend/internal/api"
	"otcountingbackend/internal/repo"
	"otcountingbackend/internal/service"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func envOrDefault(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}

func dbDSN() string {
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn
	}

	host := envOrDefault("OPENGAUSS_HOST", "localhost")
	port := envOrDefault("OPENGAUSS_PORT", "5432")
	user := envOrDefault("OPENGAUSS_USER", "postgres")
	pass := os.Getenv("OPENGAUSS_PASSWORD")
	if pass == "" {
		pass = os.Getenv("GS_PASSWORD")
	}
	if pass == "" {
		pass = "postgres"
	}
	db := envOrDefault("OPENGAUSS_DBNAME", "postgres")
	ssl := envOrDefault("OPENGAUSS_SSLMODE", "disable")

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", url.QueryEscape(user), url.QueryEscape(pass), host, port, db, url.QueryEscape(ssl))
}

func main() {
	dsn := dbDSN()
	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Printf("warning: unable to ping database at startup: %v", err)
	}

	sessionRepo := &repo.SessionRepo{DB: db}
	entryRepo := &repo.EntryRepo{DB: db}
	resultRepo := &repo.ResultRepo{DB: db}
	renderedRepo := &repo.RenderedRepo{DB: db}
	transactor := &repo.Transactor{DB: db}

	sessionSvc := &service.SessionService{Sessions: sessionRepo, Entries: entryRepo, Transactor: transactor}
	calculateSvc := &service.CalculateService{Sessions: sessionRepo, Entries: entryRepo, Results: resultRepo, Rendered: renderedRepo, Transactor: transactor, Engine: service.EngineAdapter{}}

	h := &api.Handler{SessionSvc: sessionSvc, CalculateSvc: calculateSvc, RenderedRepo: renderedRepo}
	mux := http.NewServeMux()
	h.Register(mux)

	log.Printf("server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
