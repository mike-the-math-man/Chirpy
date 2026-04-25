package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Issuer:    "chirpy-access",
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
			Subject:   userID.String(),
		})
	signed_token, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return signed_token, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		fmt.Printf("error getting token with parse %v", err)
		return uuid.Nil, err
	}
	string_id, err := token.Claims.GetSubject()
	if err != nil {
		fmt.Printf("error getting subject %v", err)
		return uuid.Nil, err
	}
	//fmt.Println(string_id)
	return uuid.Parse(string_id)
}

func GetBearerToken(headers http.Header) (string, error) {
	bearer_token := headers.Get("Authorization")
	if bearer_token == "" {
		fmt.Println("bearer_token = empty")
		return bearer_token, fmt.Errorf("No auth")
	}
	token_string := strings.Split(bearer_token, " ")[1]
	return token_string, nil
}
