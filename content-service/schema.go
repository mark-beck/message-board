package main

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type (
	UserInfo struct {
		Id    string   `json:"id"`
		Name  string   `json:"name"`
		Roles []string `json:"roles"`
		Image string   `json:"image"`
	}

	Post struct {
		Author string `json:"author" bson:"author"`
		Text   string `json:"text" bson:"text"`
		Date   string `json:"date" bson:"date"`
		Id     string `json:"id" bson:"id"`
		Image  string `json:"image" bson:"image"`
	}

	FilledPost struct {
		AuthorInfo *UserInfo `json:"author_info"`
		Post
	}

	FilledComment struct {
		AuthorInfo *UserInfo `json:"author_info"`
		Comment
	}

	Comment struct {
		Author string `json:"author" bson:"author"`
		Text   string `json:"text" bson:"text"`
		Date   string `json:"date" bson:"date"`
		Id     string `json:"id" bson:"id"`
		Parent string `json:"parent" bson:"parent"`
	}

	Filter struct {
		Author    string `json:"author" bson:"author"`
		Text      string `json:"text" bson:"text"`
		StartDate string `json:"startDate" bson:"startDate"`
		EndDate   string `json:"endDate" bson:"endDate"`
		SortBy    string `json:"sortBy" bson:"sortBy"`
		SortOrder string `json:"sortOrder" bson:"sortOrder"`
	}

	Handler struct {
		client mongo.Client
		auth   AuthServer
	}
)
