package response

// TokenResponse is the response for `ClientAuth“ and `Login` related APIs
type TokenResponse struct {
	AccessToken string `json:"accessToken"`
}
