package main

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo-contrib/jaegertracing"
	"github.com/labstack/echo/v4"
)

func (handler *Handler) get_comment(c echo.Context) error {
	sp := jaegertracing.CreateChildSpan(c, "get_comment")
	defer sp.Finish()
	id := c.Param("id")

	comment, err := mongo_get_comment(&handler.client, id)
	if err != nil {
		Error(sp, "Error retrieving document", err)
		c.Logger().Warnf("Error retrieving document: %v\n", err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	filled_comment, err := fill_comment(c, comment)
	if err != nil {
		Error(sp, "Error filling comment", err)
		c.Logger().Warnf("Error filling comment: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(200, filled_comment)
}

func (handler *Handler) get_all_comments(c echo.Context) error {
	sp := jaegertracing.CreateChildSpan(c, "get_all_comments")
	defer sp.Finish()

	comments, err := mongo_get_all_comments(&handler.client)
	if err != nil {
		Error(sp, "Error retrieving document", err)
		c.Logger().Warnf("Error retrieving document: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	filled_comments, err := fill_comments(c, comments)
	if err != nil {
		Error(sp, "Error filling comments", err)
		c.Logger().Warnf("Error filling comments: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(200, filled_comments)
}

func (handler *Handler) add_comment(c echo.Context) error {
	sp := jaegertracing.CreateChildSpan(c, "add_comment")
	defer sp.Finish()

	var comment Comment

	if err := c.Bind(&comment); err != nil {
		Error(sp, "Error binding request", err)
		c.Logger().Warnf("Error binding request: %v\n", err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	comment.Id = uuid.New().String()

	comment.Date = dateToString(time.Now())

	comment.Author = c.Get("user_id").(string)

	err := mongo_add_comment(&handler.client, comment)
	if err != nil {
		Error(sp, "Error inserting document", err)
		c.Logger().Warnf("Error inserting document: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	Info(sp, "Comment added", comment)
	c.Logger().Infof("Inserted %v", comment)
	return c.JSON(http.StatusAccepted, comment)
}

func (handler *Handler) filter_comments(c echo.Context) error {
	sp := jaegertracing.CreateChildSpan(c, "filter_comments")
	defer sp.Finish()

	var filter Filter

	if err := c.Bind(&filter); err != nil {
		Error(sp, "Error binding request", err)
		c.Logger().Warnf("Error binding request: %v\n", err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	comments, err := mongo_filter_comments(&handler.client, filter)
	if err != nil {
		Error(sp, "Error retrieving document", err)
		c.Logger().Warnf("Error retrieving document: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	filled_comments, err := fill_comments(c, comments)
	if err != nil {
		Error(sp, "Error filling comments", err)
		c.Logger().Warnf("Error filling comments: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(200, filled_comments)
}

func (handler *Handler) inject_comments(c echo.Context) error {
	sp := jaegertracing.CreateChildSpan(c, "inject_comments")
	defer sp.Finish()

	var comments []Comment

	if err := c.Bind(&comments); err != nil {
		Error(sp, "Error binding request", err)
		c.Logger().Warnf("Error binding request: %v\n", err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	for _, comment := range comments {
		if comment.Id == "" {
			comment.Id = uuid.New().String()
		}
		mongo_add_comment(&handler.client, comment)

		Info(sp, "Comment added", comment)
		c.Logger().Infof("Inserted %v", comment)
	}

	return c.String(200, "")

}

func (handler *Handler) delete_all_comments(c echo.Context) error {
	sp := jaegertracing.CreateChildSpan(c, "delete_all_comments")
	defer sp.Finish()

	err := mongo_delete_all_comments(&handler.client)
	if err != nil {
		Error(sp, "Error deleting document", err)
		c.Logger().Warnf("Error deleting document: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	Info(sp, "All comments deleted", "")
	c.Logger().Infof("Deleted all comments")

	return c.String(200, "")
}

// delete comment with id, refuse if not author or admin or moderator
func (handler *Handler) delete_comment(c echo.Context) error {
	sp := jaegertracing.CreateChildSpan(c, "delete_comment")
	defer sp.Finish()

	id := c.Param("id")

	comment, err := mongo_get_comment(&handler.client, id)
	if err != nil {
		Error(sp, "Error retrieving document", err)
		c.Logger().Warnf("Error retrieving document: %v\n", err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	user_info, err := handler.auth.get_info(c, c.Request().Header["Authorization"][0])
	if err != nil {
		Error(sp, "Error retrieving user info", err)
		c.Logger().Warnf("Error retrieving user info: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if comment.Author != user_info.Name && !user_info.isRole("admin") && !user_info.isRole("moderator") {
		Error(sp, "not authorized to delete comment", nil)
		c.Logger().Warnf("User %v is not authorized to delete comment %v", user_info.Name, id)
		return echo.NewHTTPError(http.StatusUnauthorized)
	}

	err = mongo_delete_comment(&handler.client, id)
	if err != nil {
		Error(sp, "Error deleting document", err)
		c.Logger().Warnf("Error deleting document: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	Info(sp, "Comment deleted", comment)
	c.Logger().Infof("Deleted comment %v", id)
	return c.String(200, "")
}

func fill_comment(ctx echo.Context, comment Comment) (FilledComment, error) {
	sp := jaegertracing.CreateChildSpan(ctx, "fill_comment")
	defer sp.Finish()

	info, err := init_auth().get_user(ctx, comment.Author)
	if err != nil {
		return FilledComment{}, err
	}

	return FilledComment{&info, comment}, nil
}

// fill posts with author info
func fill_comments(ctx echo.Context, comments []Comment) ([]FilledComment, error) {
	sp := jaegertracing.CreateChildSpan(ctx, "fill_posts")
	defer sp.Finish()

	infos, err := init_auth().get_user_batch(ctx, get_ids_comment(comments))
	if err != nil {
		return nil, err
	}

	var filled_comments []FilledComment

	for _, post := range comments {
		info := find_user(infos, post.Author)
		filled_comments = append(filled_comments, FilledComment{&info, post})
	}

	return filled_comments, nil
}

func get_ids_comment(posts []Comment) []string {
	var ids []string

	for _, post := range posts {
		ids = append(ids, post.Author)
	}

	return ids
}
