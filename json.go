package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

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
