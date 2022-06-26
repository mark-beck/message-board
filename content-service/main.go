package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Post struct {
	Author string `json:"author" bson:"author"`
	Text   string `json:"text" bson:"text"`
	Date   string `json:"date" bson:"date"`
	Id     string `json:"id" bson:"id"`
	Image  string `json:"image" bson:"image"`
}

type Comment struct {
	Author string `json:"author" bson:"author"`
	Text   string `json:"text" bson:"text"`
	Date   string `json:"date" bson:"date"`
	Id     string `json:"id" bson:"id"`
	Parent string `json:"parent" bson:"parent"`
}

type Filter struct {
	Author    string `json:"author" bson:"author"`
	Text      string `json:"text" bson:"text"`
	StartDate string `json:"startDate" bson:"startDate"`
	EndDate   string `json:"endDate" bson:"endDate"`
	SortBy    string `json:"sortBy" bson:"sortBy"`
	SortOrder string `json:"sortOrder" bson:"sortOrder"`
}

type Handler struct {
	client mongo.Client
	logger log.Logger
}

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
		logger: *log.Default(),
	}
	handler.client = *client

	handler.logger = *log.Default()

	r := gin.Default()

	r.Use(CORSMiddleware())
	content := r.Group("/content", AuthMiddleware(secrets))
	// content := r.Group("/content")

	{
		posts := content.Group("/posts")
		posts.GET("/all", handler.get_all_posts)
		posts.POST("/add", handler.add_post)
		posts.POST("/filter", handler.filter_posts)
		posts.POST("/inject", handler.inject_posts)
		posts.GET("/delete_all", handler.delete_all_posts)
		posts.DELETE("/delete/:id", handler.delete_post)
	}
	{
		comments := content.Group("/comments")
		comments.GET("/get/:id", handler.get_comment)
		comments.GET("/all", handler.get_all_comments)
		comments.POST("/add", handler.add_comment)
		comments.POST("/filter", handler.filter_comments)
		comments.POST("/inject", handler.inject_comments)
		comments.GET("/delete_all", handler.delete_all_comments)
		comments.DELETE("/delete/:id", handler.delete_comment)
	}

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func (userClaims UserClaims) isRole(role string) bool {
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
