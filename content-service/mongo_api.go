package main

import (
	"context"

	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func mongo_get_post(client *mongo.Client, id string) (Post, error) {
	coll := client.Database("content").Collection("posts")

	filter := bson.D{{Key: "id", Value: id}}
	opts := options.FindOne()
	post := Post{}
	err := coll.FindOne(context.TODO(), filter, opts).Decode(&post)
	return post, err
}

func mongo_get_all_posts(client *mongo.Client) ([]Post, error) {
	coll := client.Database("content").Collection("posts")

	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{Key: "date", Value: -1}})
	cursor, err := coll.Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, err
	}

	var posts []Post
	err = nil

	for cursor.Next(context.TODO()) {
		var post Post

		err = cursor.Decode(&post)
		if err != nil {
			continue
		}

		posts = append(posts, post)
	}

	return posts, err
}

func mongo_add_post(client *mongo.Client, post Post) error {
	coll := client.Database("content").Collection("posts")

	_, err := coll.InsertOne(context.TODO(), post)
	return err
}

func mongo_filter_posts(client *mongo.Client, filter Filter) ([]Post, error) {
	coll := client.Database("content").Collection("posts")
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

	log.Printf("Filter Bson: %v", filterBson)

	sortDir := -1
	if filter.SortOrder == "asc" {
		sortDir = 1
	}

	opts := options.Find().SetSort(bson.D{{Key: "date", Value: -1}})
	if filter.SortBy != "" {
		opts.SetSort(bson.D{{Key: filter.SortBy, Value: sortDir}})
	}

	log.Printf("Sort: %v", opts.Sort)

	cursor, err := coll.Find(context.TODO(), filterBson, opts)
	if err != nil {
		return nil, err
	}

	var posts []Post

	for cursor.Next(context.TODO()) {
		var post Post

		err = cursor.Decode(&post)
		if err != nil {
			log.Printf("Error decoding document: %v\n", err)
			continue
		}
		posts = append(posts, post)
	}
	return posts, err
}

func mongo_delete_all_posts(client *mongo.Client) error {
	coll := client.Database("content").Collection("posts")

	_, err := coll.DeleteMany(context.TODO(), bson.D{})
	return err
}

func mongo_delete_post(client *mongo.Client, id string) error {
	coll := client.Database("content").Collection("posts")

	_, err := coll.DeleteOne(context.TODO(), bson.D{{Key: "id", Value: id}})
	return err
}

func mongo_get_comment(client *mongo.Client, id string) (Comment, error) {
	coll := client.Database("content").Collection("comments")

	filter := bson.D{{Key: "id", Value: id}}
	opts := options.FindOne()
	comment := Comment{}
	err := coll.FindOne(context.TODO(), filter, opts).Decode(&comment)
	return comment, err
}

func mongo_get_all_comments(client *mongo.Client) ([]Comment, error) {
	coll := client.Database("content").Collection("comments")

	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{Key: "date", Value: -1}})
	cursor, err := coll.Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, err
	}

	var comments []Comment
	err = nil

	for cursor.Next(context.TODO()) {
		var comment Comment

		err = cursor.Decode(&comment)
		if err != nil {
			continue
		}

		comments = append(comments, comment)
	}

	return comments, err
}

func mongo_add_comment(client *mongo.Client, comment Comment) error {
	coll := client.Database("content").Collection("comments")

	_, err := mongo_get_post(client, comment.Parent)
	if err != nil {
		return err
	}

	_, err = coll.InsertOne(context.TODO(), comment)
	return err
}

func mongo_filter_comments(client *mongo.Client, filter Filter) ([]Comment, error) {
	coll := client.Database("content").Collection("comments")
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

	log.Printf("Filter Bson: %v", filterBson)

	sortDir := -1
	if filter.SortOrder == "asc" {
		sortDir = 1
	}

	opts := options.Find().SetSort(bson.D{{Key: "date", Value: -1}})
	if filter.SortBy != "" {
		opts.SetSort(bson.D{{Key: filter.SortBy, Value: sortDir}})
	}

	log.Printf("Sort: %v", opts.Sort)

	cursor, err := coll.Find(context.TODO(), filterBson, opts)
	if err != nil {
		return nil, err
	}

	var comments []Comment

	for cursor.Next(context.TODO()) {
		var comment Comment

		err = cursor.Decode(&comment)
		if err != nil {
			log.Printf("Error decoding document: %v\n", err)
			continue
		}
		comments = append(comments, comment)
	}
	return comments, err
}

func mongo_delete_all_comments(client *mongo.Client) error {
	coll := client.Database("content").Collection("comments")

	_, err := coll.DeleteMany(context.TODO(), bson.D{})
	return err
}

func mongo_delete_comment(client *mongo.Client, id string) error {
	coll := client.Database("content").Collection("comments")

	_, err := coll.DeleteOne(context.TODO(), bson.D{{Key: "id", Value: id}})
	return err
}
