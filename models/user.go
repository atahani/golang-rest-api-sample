package models

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)


//used for database and JSON
type User struct {
	Id             bson.ObjectId `json:"id" bson:"_id"`
	FirstName      string        `valid:"required" json:"first_name" bson:"first_name"`
	LastName       string        `valid:"required" json:"last_name" bson:"last_name"`
	DisplayName    string        `valid:"required" json:"display_name" bson:"display_name"`
	Email          string        `valid:"email,required" json:"email" bson:"email"`
	HashedPassword string        `json:"password,omitempty" bson:"hashed_password"`
	ImageFileName  string        `default:"default_image_profile.jpeg" json:"image_profile_url" bson:"image_profile_file_name"`
	IsEnable       bool          `default:"true" json:"is_enable" bson:"enable_status"`
	TrustedApps    []TrustedApp  `json:"-" bson:"trusted_apps,omitempty"`
	Roles          []string      `json:"roles" bson:"roles"`
	JoinedAt       time.Time     `json:"joined_at" bson:"joined_at"`
	UpdatedAt      time.Time     `json:"updated_at" bson:"updated_at"`
}

//only for database models
type TrustedApp struct {
	Id               bson.ObjectId      `bson:"_id"`
	ClientId         bson.ObjectId      `bson:"_client"`
	RefreshToken     string             `bson:"refresh_token"`
	DeviceModel      string             `bson:"device_model,omitempty"`
	OSVersion        string             `bson:"os_version,omitempty"`
	AppVersion       string             `bson:"app_version,omitempty"`
	MessageTokenType string             `bson:"message_token_type,omitempty"`
	MessageToken     string             `bson:"message_token,omitempty"`
	GrantedAt        time.Time          `bson:"granted_at"`
}