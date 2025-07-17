package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

var apiCfg = apiConfig{
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
	apiCfg.fileserverHits.Store(0)
	request.Header.Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(200)
}
