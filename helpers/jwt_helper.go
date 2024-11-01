package helpers

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtHelper interface {
	GenerateAccessToken(id int, name string, email string, jwtAccessTokenTime int, secret string) (accessToken string, err error)
	GenerateRefreshToken(id int, jwtRefreshTokenTime int, secret string) (refreshToken string, err error)
}

type JwtHelperImplementation struct {
}

func NewJwtHelper() JwtHelper {
	return &JwtHelperImplementation{}
}

type AccessTokenCustomClaims struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func (helper *JwtHelperImplementation) GenerateAccessToken(id int, name string, email string, jwtAccessTokenTime int, secret string) (accessToken string, err error) {
	claims := AccessTokenCustomClaims{
		id,
		name,
		email,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(jwtAccessTokenTime) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err = token.SignedString([]byte(secret))
	return
}

type refreshTokenCustomClaims struct {
	Id int
	jwt.RegisteredClaims
}

func (helper *JwtHelperImplementation) GenerateRefreshToken(id int, jwtRefreshTokenTime int, secret string) (refreshToken string, err error) {
	claims := refreshTokenCustomClaims{
		id,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(jwtRefreshTokenTime) * (time.Hour * 24))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	refreshToken, err = token.SignedString([]byte(secret))
	return
}
