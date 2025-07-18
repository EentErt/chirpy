package main

import (
	"chirpy/internal/database"
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Printf("Error: %s", err)
	}

	dbQueries := database.New(db)
	ApiCfg.Queries = *dbQueries

	ApiCfg.Platform = os.Getenv("PLATFORM")

	fmt.Println("hi")
	ApiCfg.fileserverHits.Store(0)
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	handler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))

	mux.Handle("/app/", ApiCfg.middlewareMetricsInc(handler))
	mux.HandleFunc("GET /api/healthz", healthz)
	mux.HandleFunc("GET /admin/metrics", ApiCfg.metrics)
	mux.HandleFunc("POST /admin/reset", ApiCfg.reset)
	mux.HandleFunc("POST /api/users", createUser)
	mux.HandleFunc("POST /api/chirps", createChirp)
	mux.HandleFunc("GET /api/chirps", getChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", getChirp)
	mux.HandleFunc("POST /api/login", login)

	server.ListenAndServe()
}

func healthz(writer http.ResponseWriter, request *http.Request) {
	request.Header.Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(200)
	writer.Write([]byte("OK"))
}
