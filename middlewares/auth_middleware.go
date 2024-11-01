package middlewares

import (
	"context"
	"net/http"
	"os"
	"todo-list-api/helpers"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func Authenticate(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authorizationToken, err := c.Cookie("Authorization")
		if err != nil && err != http.ErrNoCookie {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": err.Error(),
			})
		} else if err != nil && err == http.ErrNoCookie {
			return c.JSON(http.StatusNotFound, map[string]string{
				"message": "token not found",
			})
		}
		token, err := jwt.ParseWithClaims(authorizationToken.Value, &helpers.AccessTokenCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"message": "Unauthorized",
			})
		} else if claims, ok := token.Claims.(*helpers.AccessTokenCustomClaims); ok {
			ctx := context.WithValue(c.Request().Context(), "userId", claims.Id)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		} else {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "internal server error",
			})
		}
	}
}
