package models

//it's used only for JSON request
type SignInRequest struct {
	AppId       string     `valid:"required" json:"app_id"`
	AppKey      string     `json:"app_key"`
	DeviceModel string     `json:"device_model"`
	Email       string     `valid:"email,required" json:"email"`
	Password    string     `valid:"length(6|64),required" json:"password"`
}