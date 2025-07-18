package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type userRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func createUser(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	userReq := userRequest{}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		errorResponse := fmt.Sprint(err)
		respondWithJsonError(writer, errorResponse, 500)
		return
	}

	if err := json.Unmarshal(body, &userReq); err != nil {
		respondWithJsonError(writer, "Something went wrong", 500)
	}

	hashedPassword, err := auth.HashPassword(userReq.Password)
	if err != nil {
		respondWithJsonError(writer, "Something went wrong", 500)
		return
	}

	newUser := database.CreateUserParams{
		Email:          userReq.Email,
		HashedPassword: hashedPassword,
	}

	user, err := ApiCfg.Queries.CreateUser(request.Context(), newUser)
	if err != nil {
		errorResponse := fmt.Sprint(err)
		respondWithJsonError(writer, errorResponse, 500)
		return
	}

	responseBody := map[string]string{
		"id":         user.ID.String(),
		"created_at": user.CreatedAt.String(),
		"updated_at": user.UpdatedAt.String(),
		"email":      user.Email,
	}

	respondWithJson(writer, 201, responseBody)
}

func login(writer http.ResponseWriter, request *http.Request) {
	userReq := userRequest{}
	defer request.Body.Close()
	body, err := io.ReadAll(request.Body)
	if err != nil {
		respondWithJsonError(writer, "Something went wrong", 500)
		return
	}

	if err := json.Unmarshal(body, &userReq); err != nil {
		respondWithJsonError(writer, "Something went wrong", 500)
		return
	}

	user, err := ApiCfg.Queries.GetUserByEmail(request.Context(), userReq.Email)
	if err != nil {
		respondWithJsonError(writer, "Something went wrong", 500)
		return
	}

	if err := auth.CheckPasswordHash(user.HashedPassword, userReq.Password); err != nil {
		fmt.Println(err)
		respondWithJsonError(writer, "Unauthorized", 401)
		return
	}

	respondWithJson(writer, 200, makeUserMap(user))
}

func makeUserMap(user database.User) map[string]string {
	userMap := map[string]string{
		"id":         user.ID.String(),
		"created_at": user.CreatedAt.String(),
		"updated_at": user.UpdatedAt.String(),
		"email":      user.Email,
	}

	return userMap
}
