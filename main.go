package main

import (
	"fmt"
	"joho/godotenv"
	"net/http"

	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()
	fmt.Println("hi")
	apiCfg.fileserverHits.Store(0)
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	handler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))
	mux.HandleFunc("GET /api/healthz", healthz)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.reset)
	mux.HandleFunc("POST /api/validate_chirp", validateChirp)

	server.ListenAndServe()
}

func healthz(writer http.ResponseWriter, request *http.Request) {
	request.Header.Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(200)
	writer.Write([]byte("OK"))
}

type chirp struct {
	Body string `json:"body"`
}
