package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo-contrib/jaegertracing"
	"github.com/labstack/echo/v4"
)

// func CORSMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
// 		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
// 		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
// 		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

// 		if c.Request.Method == "OPTIONS" {
// 			c.AbortWithStatus(204)
// 			return
// 		}

// 		c.Next()
// 	}
// }

func (secrets *Secrets) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sp := jaegertracing.CreateChildSpan(c, "auth_middleware")
		defer sp.Finish()

		auth_header := c.Request().Header.Get("Authorization")
		if auth_header == "" || !strings.HasPrefix(auth_header, "Bearer ") {
			return echo.NewHTTPError(http.StatusUnauthorized, "missing or invalid Authorization header")
		}

		tokenString := strings.TrimPrefix(auth_header, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
			return secrets.token_public, nil
		})
		if err != nil {
			Error(sp, "Error parsing token", err)
			c.Logger().Warnf("Error while parsing token: %v\n", err)
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		}

		if !token.Valid {
			Error(sp, "Token is invalid", err)
			c.Logger().Warnf("token not valid")
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		}

		user_id := token.Claims.(jwt.MapClaims)["user_id"]

		Info(sp, "User authorized", user_id)
		c.Logger().Debugf("User ID: %v", user_id)

		c.Set("user_id", user_id)
		if err = next(c); err != nil {
			c.Error(err)
		}
		return nil
	}
}

// func AuthMiddleware(secrets Secrets) echo.HandlerFunc {
// 	return func(c *echo.Context) echo.HandlerFunc {
// 		auth_header := c.Request.Header.Get("Authorization")
// 		if auth_header == "" || !strings.HasPrefix(auth_header, "Bearer ") {
// 			c.AbortWithStatus(http.StatusUnauthorized)
// 			return
// 		}

// 		tokenString := strings.TrimPrefix(auth_header, "Bearer ")

// 		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
// 			if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
// 				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
// 			}

// 			// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
// 			return secrets.token_public, nil
// 		})
// 		if err != nil {
// 			log.Printf("Error while parsing token: %v\n", err)
// 			c.AbortWithStatus(http.StatusUnauthorized)
// 			return
// 		}

// 		if !token.Valid {
// 			log.Println("token not valid")
// 			c.AbortWithStatus(http.StatusUnauthorized)
// 			return
// 		}

// 		user_id := token.Claims.(jwt.MapClaims)["user_id"]

// 		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), "user_id", user_id))
// 		c.Next()

// 	}
// }
