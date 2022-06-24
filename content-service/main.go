package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Post struct {
	Author string `json:"author" bson:"author"`
	Text   string `json:"text" bson:"text"`
	Date   string `json:"date" bson:"date"`
	Id     string `json:"id" bson:"id"`
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
		content.GET("/all", handler.get_all)
		content.POST("/add", handler.add_content)
		content.POST("/filter", handler.get_with_filter)
		content.POST("/inject", handler.inject_posts)
		content.GET("/delete_all", handler.delete_all)
		content.DELETE("/delete/:id", handler.delete_post)
	}

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func (handler *Handler) get_all(c *gin.Context) {
	coll := handler.client.Database("content").Collection("posts")

	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{Key: "date", Value: -1}})
	cursor, err := coll.Find(context.TODO(), filter, opts)
	if err != nil {
		handler.logger.Printf("Error retrieveing documents: %v\n", err)
		c.AbortWithStatus(500)
		return
	}

	var posts []Post

	for cursor.Next(context.TODO()) {
		var post Post

		err = cursor.Decode(&post)
		if err != nil {
			handler.logger.Printf("Error decoding document: %v\n", err)
			continue
		}

		posts = append(posts, post)
	}

	c.JSON(200, posts)
}

func (handler *Handler) add_content(c *gin.Context) {

	var post Post

	coll := handler.client.Database("content").Collection("posts")

	if err := c.BindJSON(&post); err != nil {
		handler.logger.Println(err)
		c.AbortWithStatus(400)
		return
	}

	post.Id = uuid.New().String()

	post.Date = dateToString(time.Now())

	post.Author = c.Request.Context().Value("user").(UserClaims).Name

	_, err := coll.InsertOne(context.TODO(), post)
	if err != nil {
		c.AbortWithStatus(500)
	}

	handler.logger.Printf("Inserted %v", post)

}

func (handler *Handler) get_with_filter(c *gin.Context) {
	var filter Filter

	if err := c.BindJSON(&filter); err != nil {
		handler.logger.Println(err)
		c.AbortWithStatus(400)
		return
	}

	handler.logger.Printf("Filter: %v", filter)

	coll := handler.client.Database("content").Collection("posts")
	filterBson := bson.D{}
	if filter.Author != "" {
		filterBson = append(filterBson, bson.E{Key: "author", Value: filter.Author})
	}
	if filter.Text != "" {
		filterBson = append(filterBson, bson.E{
			Key: "text", Value: []bson.E{
				{Key: "$regex", Value: filter.Text},
				{Key: "$options", Value: "i"},
			},
		})
	}

	handler.logger.Printf("Filter Bson: %v", filterBson)

	sortDir := -1
	if filter.SortOrder == "asc" {
		sortDir = 1
	}

	opts := options.Find().SetSort(bson.D{{Key: "date", Value: -1}})
	if filter.SortBy != "" {
		opts.SetSort(bson.D{{Key: filter.SortBy, Value: sortDir}})
	}

	handler.logger.Printf("Sort: %v", opts.Sort)

	cursor, err := coll.Find(context.TODO(), filterBson, opts)
	if err != nil {
		handler.logger.Printf("Error retrieveing documents: %v\n", err)
		c.AbortWithStatus(500)
		return
	}

	var posts []Post

	for cursor.Next(context.TODO()) {
		var post Post

		err = cursor.Decode(&post)
		if err != nil {
			handler.logger.Printf("Error decoding document: %v\n", err)
			continue
		}
		posts = append(posts, post)
	}

	c.JSON(200, posts)
}

func (handler *Handler) inject_posts(c *gin.Context) {
	var posts []Post

	coll := handler.client.Database("content").Collection("posts")

	if err := c.BindJSON(&posts); err != nil {
		handler.logger.Println(err)
		c.AbortWithStatus(400)
		return
	}

	for _, post := range posts {
		if post.Id == "" {
			post.Id = uuid.New().String()
		}
		_, err := coll.InsertOne(context.TODO(), post)
		if err != nil {
			handler.logger.Printf("Error inserting document: %v\n", err)
			c.AbortWithStatus(500)
			return
		}

		handler.logger.Printf("Inserted %v", post)
	}

}

func (handler *Handler) delete_all(c *gin.Context) {
	coll := handler.client.Database("content").Collection("posts")

	_, err := coll.DeleteMany(context.TODO(), bson.D{})
	if err != nil {
		handler.logger.Printf("Error deleting documents: %v\n", err)
		c.AbortWithStatus(500)
		return
	}

	handler.logger.Printf("Deleted all documents")
}

// delete post with id, refuse if not author or admin or moderator
func (handler *Handler) delete_post(c *gin.Context) {
	id := c.Param("id")

	coll := handler.client.Database("content").Collection("posts")

	var post Post

	err := coll.FindOne(context.TODO(), bson.D{{Key: "id", Value: id}}).Decode(&post)

	if err != nil {
		handler.logger.Printf("Error deleting document: %v\n", err)
		c.AbortWithStatus(403)
		return
	}

	userClaims := c.Request.Context().Value("user").(UserClaims)

	if post.Author != userClaims.Name && !userClaims.isRole("admin") && !userClaims.isRole("moderator") {
		handler.logger.Printf("User %v tried to delete post %v", userClaims.Name, post)
		c.AbortWithStatus(403)
		return
	}

	_, err = coll.DeleteOne(context.TODO(), bson.D{{Key: "id", Value: id}})
	if err != nil {
		handler.logger.Printf("Error deleting document: %v\n", err)
		c.AbortWithStatus(500)
		return
	}

	handler.logger.Printf("Deleted %v", id)
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
