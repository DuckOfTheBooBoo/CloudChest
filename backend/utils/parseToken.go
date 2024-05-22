package utils

import (
	"github.com/golang-jwt/jwt"
	"os"
)

type UserClaims struct {
	ID uint `json:"id"`
	jwt.StandardClaims
}

func GenerateToken(claims UserClaims) (string, error) {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return accessToken.SignedString([]byte(os.Getenv("TOKEN_SECRET")))
}

func ParseToken(accessToken string) (*UserClaims, error) {
	parsedAccessToken, err := jwt.ParseWithClaims(accessToken, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("TOKEN_SECRET")), nil
	})

	if err != nil {
		return nil, err
	}

	return parsedAccessToken.Claims.(*UserClaims), err
}
