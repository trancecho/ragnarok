package util

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	// StandardClaims 已经弃用，使用 RegisteredClaims
	jwt.RegisteredClaims
}

// 生成 JWT Token
func GenerateToken(userID uint, username, role string, secret string, appName string) (string, error) {
	now := time.Now()
	expireTime := now.Add(24 * 7 * time.Hour) // Token 有效期 一周

	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime), // 转换为 *NumericDate
			IssuedAt:  jwt.NewNumericDate(now),        // 转换为 *NumericDate
			Issuer:    appName,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// 验证 JWT Token
func ParseToken(tokenString string, secret []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
