package models


//it's used only for JSON request
type RefreshTokenRequest struct {
	AppId        string     `valid:"required" json:"app_id"`
	AppKey       string     `json:"app_key"`
	RefreshToken string     `valid:"required" json:"refresh_token"`
}