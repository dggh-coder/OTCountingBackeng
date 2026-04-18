package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"otcountingbackend/internal/api"
	"otcountingbackend/internal/repo"
	"otcountingbackend/internal/service"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	}
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
