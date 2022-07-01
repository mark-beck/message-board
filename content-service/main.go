package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo-contrib/jaegertracing"
	"github.com/labstack/gommon/log"
	"github.com/opentracing/opentracing-go"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func main() {

	config := Config{
		token_secret: "/etc/certs/jwt.public.pem",
	}

	secrets := config.load()

	clientOpts := options.Client().ApplyURI(
		"mongodb://admin:admin@mongo:27017/?connect=direct")

	client, err := mongo.Connect(context.TODO(), clientOpts)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	err = client.Ping(context.TODO(), readpref.Primary())
	if err != nil {
		log.Fatal("cant connect to mongo instance")
	}

	handler := Handler{
		client: *client,
		auth:   init_auth(),
	}
	handler.client = *client

	e := echo.New()
	e.HideBanner = true

	e.Logger.SetLevel(log.DEBUG)
	e.Logger.SetOutput(os.Stdout)
	e.Use(middleware.Logger())

	c := jaegertracing.New(e, nil)
	defer c.Close()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))
	content := e.Group("/content", secrets.AuthMiddleware)
	// content := e.Group("/content")

	{
		posts := content.Group("/posts")
		posts.GET("/all", handler.get_all_posts)
		posts.POST("/add", handler.add_post)
		posts.POST("/filter", handler.filter_posts)
		posts.POST("/inject", handler.inject_posts)
		posts.DELETE("/all", handler.delete_all_posts)
		posts.DELETE("/:id", handler.delete_post)
	}
	{
		comments := content.Group("/comments")
		comments.GET("/:id", handler.get_comment)
		comments.GET("/all", handler.get_all_comments)
		comments.POST("/add", handler.add_comment)
		comments.POST("/filter", handler.filter_comments)
		comments.POST("/inject", handler.inject_comments)
		comments.DELETE("/all", handler.delete_all_comments)
		comments.DELETE("/:id", handler.delete_comment)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	e.Logger.Fatal(e.Start(fmt.Sprint(":" + port)))
}

func (userClaims UserInfo) isRole(role string) bool {
	for _, r := range userClaims.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// a function that converts go date objects to iso8601 strings
func dateToString(date time.Time) string {
	return date.Format("2006-01-02T15:04:05")
}

func Error(sp opentracing.Span, message string, err interface{}) {
	sp.SetTag("error", true)
	sp.LogKV("level", "ERROR", message, err)
}

func Info(sp opentracing.Span, message string, info interface{}) {
	sp.LogKV("level", "INFO", message, info)
}
