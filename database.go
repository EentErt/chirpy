package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type email struct {
	Address string `json:"email"`
}

func createUser(writer http.ResponseWriter, request *http.Request) {
	email := email{}
	body, err := io.ReadAll(request.Body)
	if err != nil {
		errorResponse := fmt.Sprint(err)
		respondWithJsonError(writer, errorResponse, 500)
	}
	defer request.Body.Close()

	json.Unmarshal(body, &email)

	user, err := ApiCfg.Queries.CreateUser(request.Context(), email.Address)
	if err != nil {
		errorResponse := fmt.Sprint(err)
		respondWithJsonError(writer, errorResponse, 500)
	}

	responseBody := map[string]string{
		"id":         user.ID.String(),
		"created_at": user.CreatedAt.String(),
		"updated_at": user.UpdatedAt.String(),
		"email":      user.Email,
	}

	respondWithJson(writer, 201, responseBody)
}
