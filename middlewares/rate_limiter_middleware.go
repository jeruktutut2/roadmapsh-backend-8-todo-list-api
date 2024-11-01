package middlewares

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

var NumberOfLimit int
var NumberOfRequest int

func SetRateLimiter(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		NumberOfRequest++
		if NumberOfRequest > NumberOfLimit {
			return c.JSON(http.StatusTooManyRequests, map[string]string{
				"message": "too many request",
			})
		}
		err := next(c)
		if err != nil {
			c.Error(err)
		}
		NumberOfRequest--
		return nil
	}
}
