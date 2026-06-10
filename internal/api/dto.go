package api

type CredentialsRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type TokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type TokenPairResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
