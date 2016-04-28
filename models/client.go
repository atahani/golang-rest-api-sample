package models

import (
	"gopkg.in/mgo.v2/bson"
	"time"
	"encoding/hex"
	"crypto/md5"
)

type Client struct {
	AppId        bson.ObjectId     `json:"app_id" bson:"_id"`
	AppKey       string            `json:"app_key" bson:"key"`
	Name         string            `valid:"required" json:"name" bson:"name"`
	Description  string            `json:"description,omitempty" bson:"description,omitempty"`
	IsEnable     bool              `default:"true" json:"is_enable" bson:"enable_status"`
	PlatformType string            `default:"web" json:"platform_type" bson:"platform_type"`
	CreatedAt    time.Time         `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at" bson:"updated_at"`
}

func (cli *Client) HashedAppKey() string {
	hasher := md5.New()
	hasher.Write([]byte(cli.AppKey))
	return hex.EncodeToString(hasher.Sum(nil))
}