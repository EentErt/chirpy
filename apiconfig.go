package main

import (
	"chirpy/internal/database"
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	Queries        database.Queries
	Platform       string
	Secret         string
	PolkaKey       string
}

var ApiCfg = apiConfig{
	fileserverHits: atomic.Int32{},
}

func (apiCfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (apiCfg *apiConfig) metrics(writer http.ResponseWriter, request *http.Request) {
	hits := int(apiCfg.fileserverHits.Load())
	output := fmt.Sprintf("<html>\n<body>\n<h1>Welcome, Chirpy Admin</h1>\n<p>Chirpy has been visited %d times!</p></body></html>", hits)
	request.Header.Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(200)
	writer.Header().Set("Content-Type", "text/html")
	writer.Write([]byte(output))
}

func (apiCfg *apiConfig) reset(writer http.ResponseWriter, request *http.Request) {
	if apiCfg.Platform != "dev" {
		respondWithJsonError(writer, "Forbidden", 403)
	}

	apiCfg.fileserverHits.Store(0)
	if err := apiCfg.Queries.Reset(request.Context()); err != nil {
		errorResponse := fmt.Sprint(err)
		respondWithJsonError(writer, errorResponse, 500)
	}
	request.Header.Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(200)
}
