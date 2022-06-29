package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (handler *Handler) get_all_posts(c *gin.Context) {
	posts, err := mongo_get_all_posts(&handler.client)
	if err != nil {
		handler.logger.Printf("Error retrieving documents: %v\n", err)
		c.AbortWithStatus(500)
		return
	}

	c.JSON(200, posts)
}

func (handler *Handler) add_post(c *gin.Context) {

	var post Post

	if err := c.BindJSON(&post); err != nil {
		handler.logger.Println(err)
		c.AbortWithStatus(400)
		return
	}

	post.Id = uuid.New().String()

	post.Date = dateToString(time.Now())

	post.Author = c.Request.Context().Value("user_id").(string)

	err := mongo_add_post(&handler.client, post)
	if err != nil {
		handler.logger.Printf("Error inserting document: %v\n", err)
		c.AbortWithStatus(500)
		return
	}

	handler.logger.Printf("Inserted %v", post)

}

func (handler *Handler) filter_posts(c *gin.Context) {
	var filter Filter

	if err := c.BindJSON(&filter); err != nil {
		handler.logger.Println(err)
		c.AbortWithStatus(400)
		return
	}

	handler.logger.Printf("Filter: %v", filter)

	posts, err := mongo_filter_posts(&handler.client, filter)
	if err != nil {
		handler.logger.Printf("Error retrieving documents: %v\n", err)
		c.AbortWithStatus(500)
		return
	}

	c.JSON(200, posts)
}

func (handler *Handler) inject_posts(c *gin.Context) {
	var posts []Post

	if err := c.BindJSON(&posts); err != nil {
		handler.logger.Println(err)
		c.AbortWithStatus(400)
		return
	}

	for _, post := range posts {
		if post.Id == "" {
			post.Id = uuid.New().String()
		}
		err := mongo_add_post(&handler.client, post)
		if err != nil {
			handler.logger.Printf("Error inserting document: %v\n", err)
			c.AbortWithStatus(500)
			return
		}

		handler.logger.Printf("Inserted %v", post)
	}

}

func (handler *Handler) delete_all_posts(c *gin.Context) {
	err := mongo_delete_all_posts(&handler.client)
	if err != nil {
		handler.logger.Printf("Error deleting documents: %v\n", err)
		c.AbortWithStatus(500)
		return
	}

	handler.logger.Printf("Deleted all documents")
}

// delete post with id, refuse if not author, admin or moderator
func (handler *Handler) delete_post(c *gin.Context) {
	id := c.Param("id")

	post, err := mongo_get_post(&handler.client, id)
	if err != nil {
		handler.logger.Printf("Error retrieving document: %v\n", err)
		c.AbortWithStatus(500)
		return
	}

	userClaims := c.Request.Context().Value("user").(UserClaims)

	if post.Author != userClaims.Name && !userClaims.isRole("admin") && !userClaims.isRole("moderator") {
		handler.logger.Printf("User %v tried to delete post %v", userClaims.Name, post)
		c.AbortWithStatus(403)
		return
	}

	err = mongo_delete_post(&handler.client, id)
	if err != nil {
		handler.logger.Printf("Error deleting document: %v\n", err)
		c.AbortWithStatus(500)
		return
	}

	handler.logger.Printf("Deleted %v", id)
}
