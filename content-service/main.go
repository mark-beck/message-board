package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Post struct {
	Author string `json:"author" bson:"author"`
	Text   string `json:"text" bson:"text"`
	Date   string `json:"date" bson:"date"`
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
	content := r.Group("/content", CORSMiddleware(), AuthMiddleware(secrets))
	{
		content.GET("/latest/:n", handler.content_latest)
		content.POST("/add", handler.add_content)
		content.OPTIONS("/add")
	}

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func (handler *Handler) content_latest(c *gin.Context) {
	n, err := strconv.ParseInt(c.Param("n"), 10, 64)
	if err != nil {
		c.AbortWithStatus(400)
		return
	}

	coll := handler.client.Database("content").Collection("posts")

	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{"date", -1}}).SetSkip(n).SetLimit(1)
	cursor, err := coll.Find(context.TODO(), filter, opts)
	if err != nil {
		handler.logger.Printf("Error retrieveing documents: %v\n", err)
		c.AbortWithStatus(500)
		return
	}

	cursor.Next(context.TODO())

	var post Post

	err = cursor.Decode(&post)
	if err != nil {
		handler.logger.Printf("Error decoding documents: %v\n", err)
		c.AbortWithStatusJSON(200, "null")
		return
	}

	c.JSON(200, post)
}

func (handler *Handler) add_content(c *gin.Context) {

	var post Post

	coll := handler.client.Database("content").Collection("posts")

	if err := c.BindJSON(&post); err != nil {
		handler.logger.Println(err)
		c.AbortWithStatus(400)
		return
	}

	t := time.Now()
	formatted := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())

	post.Date = formatted

	_, err := coll.InsertOne(context.TODO(), post)
	if err != nil {
		c.AbortWithStatus(500)
	}

	handler.logger.Printf("Inserted %v", post)

}
