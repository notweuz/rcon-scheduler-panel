package response

import "time"

type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type" example:"Bearer"`
	ExpiresIn   int64  `json:"expires_in" example:"2592000"`
}

func NewBearerToken(token string, ttl time.Duration) Token {
	return Token{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(ttl.Seconds()),
	}
}
