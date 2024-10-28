package auth

import (
	"github.com/golang-jwt/jwt/v4"
	"time"
)

// GenerateJWT generates a JWT token
func GenerateJWT() (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["authorized"] = true
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

	jwtToken, err := token.SignedString([]byte("secret"))
	if err != nil {
		return "", err
	}

	return jwtToken, nil
}
