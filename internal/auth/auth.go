package auth

import (
	"time"

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

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	})

	key, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	token.Signature = []byte(key)
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

	idString, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.UUID{}, err
	}

	id, err := uuid.FromBytes([]byte(idString))
	if err != nil {
		return uuid.UUID{}, err
	}

	return id, nil
}
