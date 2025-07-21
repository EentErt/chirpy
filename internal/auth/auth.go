package auth

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"crypto/rand"

	"github.com/google/uuid"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	pwd, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", err
	}
	return string(pwd), nil
}

func CheckPasswordHash(hashed_password, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hashed_password), []byte(password)); err != nil {
		return err
	}
	return nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Duration(3600000000))),
		Subject:   userID.String(),
	})

	key, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return key, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	}

	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, keyFunc)
	if err != nil {
		return uuid.UUID{}, err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return uuid.UUID{}, err
	}

	id, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.UUID{}, err
	}

	return id, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authSlice, ok := headers["Authorization"]
	if !ok {
		return "", fmt.Errorf("no authorization found")
	}

	auth := strings.Fields(authSlice[0])[1]
	return auth, nil
}

func MakeRefreshToken() (string, error) {
	randBytes := make([]byte, 32)
	rand.Read(randBytes)
	return hex.EncodeToString(randBytes), nil
}

func GetApiKey(headers http.Header) (string, error) {
	apiKey, ok := headers["Authorization"]
	if !ok {
		return "", fmt.Errorf("no api key found")
	}

	auth := strings.Fields(apiKey[0])[1]
	return auth, nil
}
