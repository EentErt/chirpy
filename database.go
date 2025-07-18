package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func createUser(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	newUser := database.CreateUserParams{}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		errorResponse := fmt.Sprint(err)
		respondWithJsonError(writer, errorResponse, 500)
	}

	json.Unmarshal(body, &newUser)

	hashedPass, err := auth.HashPassword(newUser.HashedPassword)
	if err != nil {
		respondWithJsonError(writer, "Something went wrong", 500)
	}
	newUser.HashedPassword = hashedPass

	user, err := ApiCfg.Queries.CreateUser(request.Context(), newUser)
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
