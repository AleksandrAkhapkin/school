package infrastruct

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/tarasova-school/internal/types"
	"net/http"
)

type CustomClaims struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
}

func ValidateJwt(tokenString string, key string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неизвестный метод подписи: %v", token.Header["alg"])
		}
		return []byte(key), nil
	})
	if err != nil {
		return token, err
	}

	if !token.Valid {
		return token, fmt.Errorf("токен не валидный, %v", token)
	}
	return token, nil
}

func (c CustomClaims) Valid() error {

	if c.UserID == 0 {
		return ErrorJWTIsBroken
	}

	if c.Role != types.RoleStudent && c.Role != types.RoleAdmin && c.Role != types.RoleTeacher {
		return ErrorJWTIsBroken
	}

	return nil
}

func GenerateJWT(userID int, role, secretKey string) (string, error) {

	claims := CustomClaims{
		UserID: userID,
		Role:   role,
	}

	tokenJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := tokenJWT.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GetClaimsByRequest(r *http.Request, secretKeyJWT string) (*CustomClaims, error) {
	tokenString := r.Header.Get("X-api-token")
	if len(tokenString) == 0 {
		return nil, ErrorJWTIsBroken
	}
	token, err := ValidateJwt(tokenString, secretKeyJWT)
	if err != nil {
		return nil, ErrorJWTIsBroken
	}
	if claims, ok := token.Claims.(*CustomClaims); ok {
		err = claims.Valid()
		if err != nil {
			return nil, err
		}
		return claims, nil
	}

	return nil, ErrorJWTIsBroken
}
