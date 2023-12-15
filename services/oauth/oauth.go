package oauth

type OAuth interface {
	RedirectLink() (string, error)
	GetProfile(code string) (*OAuthProfile, error)
}

type OAuthProfile struct {
	Name      string
	Email     string
	AccountId string
}
