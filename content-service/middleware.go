package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
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
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
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

		var userClaims UserClaims

		user := token.Claims.(jwt.MapClaims)["user"]
		userClaims.Name = user.(map[string]interface{})["name"].(string)
		userClaims.Email = user.(map[string]interface{})["email"].(string)
		userClaims.Roles = make([]string, 0)

		for _, role := range user.(map[string]interface{})["roles"].([]interface{}) {
			userClaims.Roles = append(userClaims.Roles, role.(string))
		}

		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), "user", userClaims))
		c.Next()

	}
}
