package oauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// githubOAuth implements oAuth interface
type githubOAuth struct {
	clientId string
	secret   string
}

func (g *githubOAuth) RedirectLink() (string, error) {

	params := url.Values{
		"client_id": {g.clientId},
		"scope":     {"user:email"},
	}

	endpoint := url.URL{
		Scheme:   "https",
		Host:     "github.com",
		Path:     "/login/oauth/authorize",
		RawQuery: params.Encode(),
	}

	u := endpoint.String()

	return u, nil
}

func (g *githubOAuth) GetProfile(code string) (*OAuthProfile, error) {
	// obtain access token
	type reqStruct struct {
		ClientId     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		Code         string `json:"code"`
	}
	body, err := json.Marshal(&reqStruct{
		ClientId:     g.clientId,
		ClientSecret: g.secret,
		Code:         code,
	})
	if err != nil {
		return nil, fmt.Errorf("encode body: %w", err)
	}

	resp, err := http.Post("https://github.com/login/oauth/access_token", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("post: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read resp: %w", err)
	}

	values, err := url.ParseQuery(string(respBody))
	if err != nil {
		return nil, fmt.Errorf("parse resp: %w", err)
	}

	if values.Get("error") != "" {
		return nil, fmt.Errorf("github api: %s", values.Get("error"))
	}

	accessToken := values.Get("access_token")

	// user profile
	profile := OAuthProfile{}

	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)

	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	defer resp2.Body.Close()

	type resp2Struct struct {
		Login string `json:"login"`
		Id    int    `json:"id"`
	}

	var data resp2Struct

	err = json.NewDecoder(resp2.Body).Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("decode resp: %w", err)
	}

	if data.Login == "" || data.Id == 0 {
		return nil, fmt.Errorf("empty user name or user id")
	}

	profile.AccountId = fmt.Sprintf("%d", data.Id)
	profile.Name = data.Login

	// email
	req2, err := http.NewRequest(http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	req2.Header.Add("Authorization", "Bearer "+accessToken)

	resp3, err := http.DefaultClient.Do(req2)
	if err != nil {
		return nil, fmt.Errorf("get email: %w", err)
	}
	defer resp3.Body.Close()

	type resp3Struct struct {
		Email    string `json:"email"`
		Verified bool   `json:"verified"`
		Primary  bool   `json:"primary"`
	}

	var data3 []resp3Struct

	err = json.NewDecoder(resp3.Body).Decode(&data3)
	if err != nil {
		return nil, fmt.Errorf("decode email json: %w", err)
	}

	for _, email := range data3 {
		if email.Primary && email.Verified {
			profile.Email = strings.ToLower(email.Email)
			return &profile, nil
		}
	}

	return nil, fmt.Errorf("primary email unverified")
}

func NewGithubOAuth(clientId string, secret string) *githubOAuth {
	return &githubOAuth{
		clientId: clientId,
		secret:   secret,
	}
}
