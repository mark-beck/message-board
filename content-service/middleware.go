package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "*")
		c.Header("Access-Control-Allow-Headers", "*")
		c.Header("Content-Type", "application/json")

		// Second, we handle the OPTIONS problem
		if c.Request.Method != "OPTIONS" {

			c.Next()

		} else {

			// Everytime we receive an OPTIONS request,
			// we just return an HTTP 200 Status Code
			// Like this, Angular can now do the real
			// request using any other method than OPTIONS
			c.AbortWithStatus(http.StatusOK)
		}
	}
}

func AuthMiddleware(secrets Secrets) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth_header := c.Request.Header.Get("Authorization")
		if auth_header == "" || !strings.HasPrefix(auth_header, "Bearer ") {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(auth_header, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
			return secrets.token_public, nil
		})
		if err != nil {
			log.Printf("Errorwhile parsing token: %v\n", err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			log.Println("token not valid")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()

	}
}
