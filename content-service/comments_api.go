package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (handler *Handler) get_comment(c *gin.Context) {
	id := c.Param("id")

	comment, err := mongo_get_comment(&handler.client, id)
	if err != nil {
		handler.logger.Printf("Error retrieving document: %v\n", err)
		c.AbortWithStatus(500)
		return
	}

	c.JSON(200, comment)
}

func (handler *Handler) get_all_comments(c *gin.Context) {
	comments, err := mongo_get_all_comments(&handler.client)
	if err != nil {
		handler.logger.Printf("Error retrieving document: %v\n", err)
		c.AbortWithStatus(500)
		return
	}

	c.JSON(200, comments)
}

func (handler *Handler) add_comment(c *gin.Context) {

	var comment Comment

	if err := c.BindJSON(&comment); err != nil {
		handler.logger.Println(err)
		c.AbortWithStatus(400)
		return
	}

	comment.Id = uuid.New().String()

	comment.Date = dateToString(time.Now())

	comment.Author = c.Request.Context().Value("user").(UserClaims).Name

	err := mongo_add_comment(&handler.client, comment)
	if err != nil {
		handler.logger.Printf("Error inserting document: %v\n", err)
		c.AbortWithStatus(500)
		return
	}

	handler.logger.Printf("Inserted %v", comment)
}

func (handler *Handler) filter_comments(c *gin.Context) {
	var filter Filter

	if err := c.BindJSON(&filter); err != nil {
		handler.logger.Println(err)
		c.AbortWithStatus(400)
		return
	}

	comments, err := mongo_filter_comments(&handler.client, filter)
	if err != nil {
		handler.logger.Printf("Error retrieving document: %v\n", err)
		c.AbortWithStatus(500)
		return
	}

	c.JSON(200, comments)
}

func (handler *Handler) inject_comments(c *gin.Context) {
	var comments []Comment

	if err := c.BindJSON(&comments); err != nil {
		handler.logger.Println(err)
		c.AbortWithStatus(400)
		return
	}

	for _, comment := range comments {
		if comment.Id == "" {
			comment.Id = uuid.New().String()
		}
		mongo_add_comment(&handler.client, comment)

		handler.logger.Printf("Inserted %v", comment)
	}

}

func (handler *Handler) delete_all_comments(c *gin.Context) {
	err := mongo_delete_all_comments(&handler.client)
	if err != nil {
		handler.logger.Printf("Error deleting document: %v\n", err)
		c.AbortWithStatus(500)
		return
	}

	handler.logger.Printf("Deleted all documents")
}

// delete post with id, refuse if not author or admin or moderator
func (handler *Handler) delete_comment(c *gin.Context) {
	id := c.Param("id")

	comment, err := mongo_get_comment(&handler.client, id)
	if err != nil {
		handler.logger.Printf("Error retrieving document: %v\n", err)
		c.AbortWithStatus(500)
		return
	}

	userClaims := c.Request.Context().Value("user").(UserClaims)

	if comment.Author != userClaims.Name && !userClaims.isRole("admin") && !userClaims.isRole("moderator") {
		handler.logger.Printf("User %v tried to delete post %v", userClaims.Name, comment)
		c.AbortWithStatus(403)
		return
	}

	err = mongo_delete_comment(&handler.client, id)
	if err != nil {
		handler.logger.Printf("Error deleting document: %v\n", err)
		c.AbortWithStatus(500)
		return
	}

	handler.logger.Printf("Deleted %v", id)
}
