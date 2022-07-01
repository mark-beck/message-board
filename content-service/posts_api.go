package main

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo-contrib/jaegertracing"
	"github.com/labstack/echo/v4"
)

func (handler *Handler) get_all_posts(c echo.Context) error {
	sp := jaegertracing.CreateChildSpan(c, "get_all_posts")
	defer sp.Finish()

	posts, err := mongo_get_all_posts(&handler.client)
	if err != nil {
		Error(sp, "Error retrieving documents", err)
		c.Logger().Warnf("Error retrieving documents: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error retrieving documents")
	}

	return c.JSON(http.StatusOK, posts)
}

func (handler *Handler) add_post(c echo.Context) error {
	sp := jaegertracing.CreateChildSpan(c, "add_post")
	defer sp.Finish()

	var post Post

	if err := c.Bind(&post); err != nil {
		Error(sp, "Error binding request", err)
		c.Logger().Printf("Error binding request: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error binding post")
	}

	post.Id = uuid.New().String()

	post.Date = dateToString(time.Now())

	post.Author = c.Get("user_id").(string)

	err := mongo_add_post(&handler.client, post)
	if err != nil {
		Error(sp, "Error inserting document", err)
		c.Logger().Printf("Error inserting document: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error inserting document")
	}

	Info(sp, "Inserted", post)
	c.Logger().Infof("Inserted %v", post)
	return c.JSON(http.StatusAccepted, post)

}

func (handler *Handler) filter_posts(c echo.Context) error {
	sp := jaegertracing.CreateChildSpan(c, "filter_posts")
	defer sp.Finish()

	var filter Filter

	if err := c.Bind(&filter); err != nil {
		Error(sp, "Error binding request", err)
		c.Logger().Printf("Error binding request: %v\n", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Error binding filter")
	}

	Info(sp, "Filter", filter)
	c.Logger().Infof("Filter: %v", filter)

	posts, err := mongo_filter_posts(&handler.client, filter)
	if err != nil {
		Error(sp, "Error retrieving documents", err)
		c.Logger().Warnf("Error retrieving documents: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error retrieving documents")
	}

	return c.JSON(200, posts)
}

func (handler *Handler) inject_posts(c echo.Context) error {
	sp := jaegertracing.CreateChildSpan(c, "inject_posts")
	defer sp.Finish()

	var posts []Post

	if err := c.Bind(&posts); err != nil {
		Error(sp, "Error binding request", err)
		c.Logger().Warn(err)
		return echo.NewHTTPError(http.StatusBadRequest, "Error binding posts")
	}

	for _, post := range posts {
		if post.Id == "" {
			post.Id = uuid.New().String()
		}
		err := mongo_add_post(&handler.client, post)
		if err != nil {
			Error(sp, "Error inserting document", err)
			c.Logger().Warnf("Error inserting document: %v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Error inserting document")
		}

		Info(sp, "Inserted", post)
		c.Logger().Infof("Inserted %v", post)
	}

	return c.String(http.StatusAccepted, "")
}

func (handler *Handler) delete_all_posts(c echo.Context) error {
	sp := jaegertracing.CreateChildSpan(c, "delete_all_posts")
	defer sp.Finish()

	err := mongo_delete_all_posts(&handler.client)
	if err != nil {
		Error(sp, "Error deleting documents", err)
		c.Logger().Warnf("Error deleting documents: %v\n", err)
		return c.String(http.StatusInternalServerError, "Error deleting documents")
	}

	Info(sp, "Deleted all documents", "")
	c.Logger().Warnf("Deleted all documents")
	return c.String(http.StatusAccepted, "")
}

// delete post with id, refuse if not author, admin or moderator
func (handler *Handler) delete_post(c echo.Context) error {
	sp := jaegertracing.CreateChildSpan(c, "delete_post")
	defer sp.Finish()

	id := c.Param("id")

	post, err := mongo_get_post(&handler.client, id)
	if err != nil {
		Error(sp, "Error retrieving document", err)
		c.Logger().Warnf("Error retrieving document: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error retrieving document")
	}

	user_info, err := handler.auth.get_info(c.Request().Header["Authorization"][0])
	if err != nil {
		Error(sp, "Error getting user info", err)
		c.Logger().Warnf("Error retrieving user info: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error retrieving user info")
	}

	if post.Author != user_info.Name && !user_info.isRole("admin") && !user_info.isRole("moderator") {
		Error(sp, "User is not authorized to delete post", user_info)
		c.Logger().Warnf("User %v is not authorized to delete post %v", user_info.Email, id)
		return echo.NewHTTPError(http.StatusUnauthorized, "User is not authorized to delete post")
	}

	err = mongo_delete_post(&handler.client, id)
	if err != nil {
		Error(sp, "Error deleting document", err)
		c.Logger().Warnf("Error deleting document: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "")
	}

	Info(sp, "Deleted", id)
	c.Logger().Infof("Deleted %v", id)
	return c.String(http.StatusAccepted, "")
}
