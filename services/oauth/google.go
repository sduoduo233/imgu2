package oauth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// googleOAuth implements oAuth interface
type googleOAuth struct {
	clientId string
	secret   string
	siteUrl  string
}

func (g *googleOAuth) RedirectLink() (string, error) {
	params := url.Values{
		"client_id":     {g.clientId},
		"redirect_uri":  {g.siteUrl + "/login/google/callback"},
		"response_type": {"code"},
		"scope":         {"https://www.googleapis.com/auth/userinfo.profile https://www.googleapis.com/auth/userinfo.email"},
	}

	endpoint := url.URL{
		Scheme:   "https",
		Host:     "accounts.google.com",
		Path:     "/o/oauth2/v2/auth",
		RawQuery: params.Encode(),
	}

	u := endpoint.String()

	return u, nil
}

func (g *googleOAuth) GetProfile(code string) (*OAuthProfile, error) {
	type respStruct struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}

	// obtian access token
	form := url.Values{
		"client_id":     {g.clientId},
		"client_secret": {g.secret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {g.siteUrl + "/login/google/callback"},
	}
	resp, err := http.PostForm("https://oauth2.googleapis.com/token", form)
	if err != nil {
		return nil, fmt.Errorf("post: %w", err)
	}
	defer resp.Body.Close()

	var data respStruct

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}

	if data.Error != "" {
		return nil, fmt.Errorf("oauth: %s", data.Error)
	}

	// user profile
	resp2, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + data.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}
	defer resp2.Body.Close()

	type profileStruct struct {
		Id            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
	}

	var profile profileStruct

	err = json.NewDecoder(resp2.Body).Decode(&profile)
	if err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}

	if profile.Id == "" || profile.Email == "" {
		return nil, fmt.Errorf("empty id or email")
	}

	if !profile.VerifiedEmail {
		return nil, fmt.Errorf("unverified email")
	}

	return &OAuthProfile{
		Email:     profile.Email,
		AccountId: profile.Id,
		Name:      profile.Name,
	}, nil
}

func NewGoogleOAuth(clientId string, secret string, siteUrl string) *googleOAuth {
	return &googleOAuth{
		clientId: clientId,
		secret:   secret,
		siteUrl:  siteUrl,
	}
}
