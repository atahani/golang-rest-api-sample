package models

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Article struct {
	Id        bson.ObjectId        `json:"id" bson:"_id"`
	Title     string               `valid:"required" json:"title" bson:"title"`
	Content   string               `valid:"required" json:"content" bson:"content"`
	UserId    bson.ObjectId        `json:"user_id" bson:"user_id"`
	CreatedAt time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time            `json:"updated_at" bson:"updated_at"`
}