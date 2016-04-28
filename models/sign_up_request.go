package models

//it's used only for JSON request
type SignUpRequest struct {
	AppId       string     `valid:"required" json:"app_id"`
	AppKey      string     `json:"app_key"`
	DeviceModel string     `json:"device_model"`
	FirstName   string     `valid:"required" json:"first_name" bson:"first_name"`
	LastName    string     `valid:"required" json:"last_name" bson:"last_name"`
	DisplayName string     `valid:"required" json:"display_name" bson:"display_name"`
	Email       string     `valid:"email,required" json:"email"`
	Password    string     `valid:"length(6|64),required" json:"password"`
}