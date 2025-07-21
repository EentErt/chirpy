package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
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
	jwtKey, err := auth.MakeJWT(user.ID, ApiCfg.Secret)
	if err != nil {
		respondWithJsonError(writer, "unable to create login key", 500)
		return
	}

	refToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithJsonError(writer, "unable to create refresh token", 500)
		return
	}

	params := database.MakeRefreshTokenParams{
		Token:  refToken,
		UserID: user.ID,
	}
	if err = ApiCfg.Queries.MakeRefreshToken(request.Context(), params); err != nil {
		respondWithJsonError(writer, "Something went wrong", 500)
	}

	userMap["token"] = jwtKey
	userMap["refresh_token"] = refToken
	respondWithJson(writer, 200, userMap)
}

func makeUserMap(user database.User) map[string]interface{} {
	userMap := map[string]interface{}{
		"id":            user.ID.String(),
		"created_at":    user.CreatedAt.String(),
		"updated_at":    user.UpdatedAt.String(),
		"email":         user.Email,
		"is_chirpy_red": user.IsChirpyRed,
	}

	return userMap
}

func refresh(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	authorization, err := auth.GetBearerToken(request.Header)
	if err != nil {
		respondWithJsonError(writer, "No authorization found", 401)
		return
	}

	token, err := ApiCfg.Queries.CheckRefreshToken(request.Context(), authorization)
	if err != nil {
		respondWithJsonError(writer, "Unauthorized", 401)
		return
	}
	if token.ExpiresAt.Before(time.Now()) {
		respondWithJsonError(writer, "Unauthorized", 401)
		return
	}
	if token.RevokedAt.Valid && token.RevokedAt.Time.Before(time.Now()) {
		respondWithJsonError(writer, "Unauthorized", 401)
		return
	}

	user, err := ApiCfg.Queries.GetUserFromRefreshToken(request.Context(), token.Token)
	if err != nil {
		respondWithJsonError(writer, "Something went wrong", 500)
		return
	}

	jwt, err := auth.MakeJWT(user.ID, ApiCfg.Secret)
	if err != nil {
		respondWithJsonError(writer, "Something went wrong", 500)
		return
	}

	respondWithJson(writer, 200, map[string]string{"token": jwt})
}

func revoke(writer http.ResponseWriter, request *http.Request) {
	token, err := auth.GetBearerToken(request.Header)
	if err != nil {
		respondWithJsonError(writer, "Unauthorized", 401)
		return
	}

	if err = ApiCfg.Queries.RevokeRefreshToken(request.Context(), token); err != nil {
		respondWithJsonError(writer, "Something went wrong", 500)
		return
	}

	writer.WriteHeader(204)
}

func updateUser(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	userReq := userRequest{}

	//get the user's authentication from the request header
	token, err := auth.GetBearerToken(request.Header)
	if err != nil {
		respondWithJsonError(writer, "Unauthorized", 401)
		return
	}

	// verify the authentication
	userID, err := auth.ValidateJWT(token, ApiCfg.Secret)
	if err != nil {
		respondWithJsonError(writer, "Unauthorized", 401)
		return
	}

	// read the request body
	body, err := io.ReadAll(request.Body)
	if err != nil {
		respondWithJsonError(writer, "Something went wrong", 500)
		return
	}

	// unmarshal the body json
	if err := json.Unmarshal(body, &userReq); err != nil {
		respondWithJsonError(writer, "Something went wrong", 500)
		return
	}

	// hash the password
	hashedPassword, err := auth.HashPassword(userReq.Password)
	if err != nil {
		respondWithJsonError(writer, "Something went wrong", 500)
		return
	}

	// set the argument parameters for UpdateUser
	params := database.UpdateUserParams{
		Email:          userReq.Email,
		HashedPassword: hashedPassword,
		ID:             userID,
	}

	// add entry to database
	user, err := ApiCfg.Queries.UpdateUser(request.Context(), params)
	if err != nil {
		respondWithJsonError(writer, "Something went wrong", 500)
		return
	}

	respondWithJson(writer, 200, makeUserMap(user))
}

type upgradeUserRequest struct {
	Event string `json:"event"`
	Data  struct {
		UserID string `json:"user_id"`
	} `json:"data"`
}

func upgradeUser(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	upgradeReq := upgradeUserRequest{}

	//check the api key
	apiKey, err := auth.GetApiKey(request.Header)
	if err != nil {
		respondWithJsonError(writer, "Unauthorized", 401)
		return
	}
	if apiKey != ApiCfg.PolkaKey {
		respondWithJsonError(writer, "Unauthorized", 401)
		return
	}

	// read request body
	body, err := io.ReadAll(request.Body)
	if err != nil {
		respondWithJsonError(writer, "Something went wrong", 500)
		return
	}

	// unmarshal body json
	if err := json.Unmarshal(body, &upgradeReq); err != nil {
		respondWithJsonError(writer, "Something went wrong", 500)
		return
	}

	// if request is invalid, return with status 204
	if upgradeReq.Event == "" || upgradeReq.Data.UserID == "" {
		respondWithJson(writer, 204, "")
		return
	}

	userID, err := uuid.Parse(upgradeReq.Data.UserID)
	if err != nil {
		respondWithJsonError(writer, "Something went wrong", 500)
		return
	}

	// if request is valid, upgrade user
	if err := ApiCfg.Queries.UpgradeUser(request.Context(), userID); err != nil {
		respondWithJsonError(writer, "User not found", 404)
		return
	}

	respondWithJson(writer, 204, "")
}
