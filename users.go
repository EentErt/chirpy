package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type userRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	Expiration int64  `json:"expires_in_seconds"`
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

	respondWithJson(writer, 201, makeUserMap(user))
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

	if userReq.Expiration == 0 || userReq.Expiration > 3600 {
		userReq.Expiration = 3600
	}
	expiration := time.Duration(userReq.Expiration * 1000000000)

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

	userMap := makeUserMap(user)
	jwtKey, err := auth.MakeJWT(user.ID, ApiCfg.Secret, expiration)
	if err != nil {
		respondWithJsonError(writer, "unable to create login key", 500)
		return
	}
	userMap["token"] = jwtKey

	respondWithJson(writer, 200, userMap)
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
