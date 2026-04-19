package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)


var ErrParseToken = errors.New("failed to parse token")

type Claims struct{
	jwt.RegisteredClaims
	UserId int
}

func ParseJwtToken (jwtString string, secretKey string) (int, error) {
	claims:= &Claims{}

	token, err:=jwt.ParseWithClaims(jwtString,claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		return 0 ,fmt.Errorf("parse token: %w",err)
	}

	if !token.Valid {
		return 0 ,fmt.Errorf("parse token: %w",ErrParseToken)
	}

	return  claims.UserId, nil
}

func BuildJwtToken (userId int, secretKey string) (string,error) {
	token:=jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour*24)), 
		},
		UserId: userId,
	})

	signed,err:=token.SignedString([]byte(secretKey))

	if err !=nil {
		return "",fmt.Errorf("signing token error: %w", err)
	}

	return signed,nil
}