package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	SecretKey []byte
}

func NewJWTManager() *JWTManager {
	j := &JWTManager{
		SecretKey: []byte("your-secret-key"),
	}
	return j
}

/*
CreateJWTTokenString метод создающий токен для авторизации
*/
func (j *JWTManager) CreateJWTToken(userName string) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"name": userName,
		"nbf":  now.Unix(),
		"exp":  now.Add(5 * time.Minute).Unix(),
		"iat":  now.Unix(),
	})

	tokenString, err := token.SignedString(j.SecretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

/*
ValidateJWTToken метод проверяющий валидность токена
*/
func (j *JWTManager) ValidateJWTToken(tokenString string) (string, error) {
	tokenFromString, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return j.SecretKey, nil
	})

	if err != nil {
		return "", err
	}

	claims, ok := tokenFromString.Claims.(jwt.MapClaims)
	if !ok || !tokenFromString.Valid {
		return "", fmt.Errorf("invalid token")
	} else {
		return claims["name"].(string), nil
	}
}
