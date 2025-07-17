package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
)

func validateChirp(writer http.ResponseWriter, request *http.Request) {
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

	respondWithJson(writer, 200, map[string]string{"cleaned_body": censorChirp(chirp.Body)})
}

func respondWithJsonError(writer http.ResponseWriter, body string, status int) {
	writer.WriteHeader(status)
	writer.Header().Set("Content-Type", "text/json")
	jsonBody := fmt.Sprintf("{\"error\": \"%s\"}", body)
	writer.Write([]byte(jsonBody))
}

func respondWithJson(writer http.ResponseWriter, status int, payload interface{}) {
	writer.WriteHeader(status)
	writer.Header().Set("Content-Type", "text/json")
	response, err := json.Marshal(payload)
	if err != nil {
		respondWithJsonError(writer, "Something went wrong", 500)
	}
	writer.Write([]byte(response))
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
