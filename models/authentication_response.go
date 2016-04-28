package models

//it's used only for JSON response in authentication and refresh token requests
type AuthenticationResponse struct {
	TokenType    string     `json:"token_type"`
	AccessToken  string     `json:"access_token"`
	ExpiresInMin float64    `json:"expire_in_min"`
	RefreshToken string     `json:"refresh_token"`
}
