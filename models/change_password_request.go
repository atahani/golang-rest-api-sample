package models

type ChangePasswordRequestModel struct {
	OldPassword string `valid:"length(6|64),required" json:"old_password"`
	Password    string `valid:"length(6|64),required" json:"password"`
}