package main

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type Post struct {
	author string
	text   string
}

type Handler struct {
	posts []Post
}

func main() {
	handler := new(Handler)

	handler.posts = make([]Post, 0)

	r := gin.Default()
	content := r.Group("/content")
	{
		content.GET("/latest/:n", handler.content_latest)
		content.POST("/add", handler.add_content)
	}

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func (handler *Handler) content_latest(c *gin.Context) {
	n, err := strconv.Atoi(c.Param("n"))
	if err != nil {
		c.AbortWithStatus(400)
		return
	}

	if len(handler.posts) <= n {
		c.AbortWithStatusJSON(200, nil)
		return
	}

	post := handler.posts[n]

	c.JSON(200, post)
}

func (handler *Handler) add_content(c *gin.Context) {

	var post Post

	if err := c.BindJSON(post); err != nil {
		c.AbortWithStatus(400)
		return
	}

	handler.posts = append(handler.posts, post)
}
