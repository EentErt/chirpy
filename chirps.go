package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"sort"
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
	defer request.Body.Close()
	chirp := chirp{}
	body, err := io.ReadAll(request.Body)
	if err != nil {
		respondWithJsonError(writer, "Something went wrong", 500)
		return
	}

	if err := json.Unmarshal(body, &chirp); err != nil {
		respondWithJsonError(writer, "Something went wrong", 500)
		return
	}

	if len(chirp.Body) > 140 {
		respondWithJsonError(writer, "Chirp is too long", 400)
		return
	}

	token, err := auth.GetBearerToken(request.Header)
	if err != nil {
		respondWithJsonError(writer, "Unauthorized", 401)
		return
	}

	id, err := auth.ValidateJWT(token, ApiCfg.Secret)
	if err != nil {
		respondWithJsonError(writer, "Unauthorized", 401)
		return
	}

	chirpParams := database.CreateChirpParams{
		Body:   censorChirp(chirp.Body),
		UserID: id,
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
	authorID := request.URL.Query().Get("author_id")
	sortReq := request.URL.Query().Get("sort")
	var chirps []database.Chirp
	var err error

	if authorID == "" {
		chirps, err = ApiCfg.Queries.GetChirps(request.Context())
	} else {
		authorUUID, errParse := uuid.Parse(authorID)
		if errParse != nil {
			respondWithJsonError(writer, "Something went wrong", 500)
			return
		}

		chirps, err = ApiCfg.Queries.GetChirpsByUser(request.Context(), authorUUID)
	}

	if err != nil {
		respondWithJsonError(writer, "Something went wrong", 500)
		return
	}

	if sortReq == "desc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		})
	}
	respondWithJson(writer, 200, makeChirpsSlice(chirps))
}

func getChirp(writer http.ResponseWriter, request *http.Request) {
	idString := request.PathValue("chirpID")
	chirpID, err := uuid.Parse(idString)
	if err != nil {
		respondWithJsonError(writer, "Something went wrong", 500)
		return
	}

	chirp, err := ApiCfg.Queries.GetChirp(request.Context(), chirpID)
	if err != nil {
		respondWithJsonError(writer, "Chirp not found", 404)
		return
	}

	respondWithJson(writer, 200, makeChirpMap(chirp))
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

func deleteChirp(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	// get the chirp id from the url
	idString := request.PathValue("chirpID")
	chirpID, err := uuid.Parse(idString)
	if err != nil {
		respondWithJsonError(writer, "Something went wrong", 500)
		return
	}

	// get the token from the header
	authorization, err := auth.GetBearerToken(request.Header)
	if err != nil {
		respondWithJsonError(writer, "Unauthorized", 401)
		return
	}

	// verify the token
	userID, err := auth.ValidateJWT(authorization, ApiCfg.Secret)
	if err != nil {
		respondWithJsonError(writer, "Unauthorized", 401)
		return
	}

	// get the chirp
	chirp, err := ApiCfg.Queries.GetChirp(request.Context(), chirpID)
	if err != nil {
		respondWithJsonError(writer, "Chirp not found", 404)
		return
	}

	// check that the user is the author of the chirp
	if chirp.UserID != userID {
		respondWithJsonError(writer, "Forbidden", 403)
		return
	}

	// delete the chirp
	ApiCfg.Queries.DeleteChirp(request.Context(), chirpID)
	respondWithJson(writer, 204, "Chirp deleted")
}
