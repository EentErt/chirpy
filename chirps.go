package main

import (
	"chirpy/internal/database"
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

type chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func createChirp(writer http.ResponseWriter, request *http.Request) {
	chirp := chirp{}
	body, err := io.ReadAll(request.Body)
	if err != nil {
		respondWithJsonError(writer, "Something went wrong", 500)
		return
	}
	defer request.Body.Close()

	if err := json.Unmarshal(body, &chirp); err != nil {
		respondWithJsonError(writer, "Something went wrong", 500)
		return
	}

	if len(chirp.Body) > 140 {
		respondWithJsonError(writer, "Chirp is too long", 400)
		return
	}

	chirpParams := database.CreateChirpParams{
		Body:   censorChirp(chirp.Body),
		UserID: chirp.UserID,
	}

	chirpData, err := ApiCfg.Queries.CreateChirp(request.Context(), chirpParams)
	if err != nil {
		respondWithJsonError(writer, "Something went wrong", 500)
		return
	}

	respondWithJson(writer, 201, makeChirpMap(chirpData))
	request.Body.Close()
}

func getChirps(writer http.ResponseWriter, request *http.Request) {
	chirps, err := ApiCfg.Queries.GetChirps(request.Context())
	if err != nil {
		respondWithJsonError(writer, "Something went wrong", 500)
		return
	}

	respondWithJson(writer, 200, makeChirpsSlice(chirps))
}

func censorChirp(body string) string {
	wordList := strings.Fields(body)
	bannedWords := []string{"kerfuffle", "sharbert", "fornax"}
	for i, word := range wordList {
		if slices.Contains(bannedWords, strings.ToLower(word)) {
			wordList[i] = "****"
		}
	}
	return strings.Join(wordList, " ")
}

func makeChirpsSlice(chirps []database.Chirp) []map[string]string {
	chirpsSlice := []map[string]string{}

	for _, chirp := range chirps {
		chirpsSlice = append(chirpsSlice, makeChirpMap(chirp))
	}

	return chirpsSlice
}

func makeChirpMap(chirp database.Chirp) map[string]string {
	chirpMap := map[string]string{
		"id":         chirp.ID.String(),
		"created_at": chirp.CreatedAt.String(),
		"updated_at": chirp.UpdatedAt.String(),
		"body":       chirp.Body,
		"user_id":    chirp.UserID.String(),
	}

	return chirpMap
}
