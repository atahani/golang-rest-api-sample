package models

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type AccessToken struct {
	Id           bson.ObjectId    `bson:"_id"`
	UserId       bson.ObjectId    `bson:"user_id"`
	TrustedAppId bson.ObjectId    `bson:"trusted_app_id"`
	Token        string           `bson:"token"`
	ExpireAt     time.Time        `bson:"expire_at"`
}