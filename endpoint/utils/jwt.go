package utils

import (
	"crud-app/models"
	"time"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("a_very_long_secreate_key_that_is_a_secreat_key_i_think")

func GenerateToken(user models.user) (string, error) {
	claims := modelsJWTClaims{
		UserID: user.ID,
		Nip: user.Nip,
		Role: user.Role,
		RegisteredClaims: jwt.RegisterdClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
		
		

	
}

func ValidateToken(tokenString string)(*models.JWTClaims, error){
		token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{},
		func(token *jwt.Token)(interface{}, error){
			return jwtSecret, nil
		})
		if err != nil{
			return nil, err
		}

		if claims, ok := token.Claims.(*models.JWTClaims); ok && token.Valid{
			teturn claims, nil
		}

		return nil , jwt.ErrInvalidKey
}
