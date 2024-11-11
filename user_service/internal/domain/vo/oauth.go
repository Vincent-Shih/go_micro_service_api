package vo

import "go_micro_service_api/pkg/enum"

type OAuthSession struct {
	Provider    enum.Provider
	AccessToken string
}

func NewOAuthSession(provider enum.Provider, accessToken string) *OAuthSession {
	return &OAuthSession{
		Provider:    provider,
		AccessToken: accessToken,
	}
}
