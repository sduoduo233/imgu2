package services

import (
	"encoding/json"
	"fmt"
	"img2/db"
	"img2/utils"
	"log/slog"
	"net/http"
	"net/url"

	"golang.org/x/crypto/bcrypt"
)

const (
	RoleAdmin  = 0
	RoleUser   = 1
	RoleBanned = 2

	// social login types

	SocialLoginGoogle = "google"
	SocialLoginGithub = "github"
)

type user struct{}

var User = user{}

// user profile obtained from oauth services (e.g. google user profile)
type OAuthProfile struct {
	Name      string
	Email     string
	AccountId string
}

func (user) FindByEmail(email string) (*db.User, error) {
	return db.UserFindByEmail(email)
}

func (user) ChangePassword(id int, currentPasswd string, newPasswd string) error {
	user, err := db.UserFindById(id)
	if err != nil {
		return err
	}

	if user == nil {
		return fmt.Errorf("user not found: %d", id)
	}

	if user.Password == "" {
		// password not set
		hashed, err := bcrypt.GenerateFromPassword([]byte(newPasswd), 0)
		if err != nil {
			return fmt.Errorf("bcrypt: %w", err)
		}
		err = db.UserChangePassword(id, string(hashed))
		if err != nil {
			return fmt.Errorf("change password: %w", err)
		}
		return nil
	}

	// compare current password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPasswd))
	if err != nil {
		return fmt.Errorf("current password does not match")
	}

	// change password
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPasswd), 0)
	if err != nil {
		return fmt.Errorf("bcrypt: %w", err)
	}
	err = db.UserChangePassword(id, string(hashed))
	if err != nil {
		return fmt.Errorf("change password: %w", err)
	}

	return nil
}

// login the user using email and password
//
// return a new token
func (user) Login(email string, password string) (string, error) {
	user, err := User.FindByEmail(email)
	if err != nil {
		slog.Error("login", "err", err)
		return "", fmt.Errorf("login: %w", err)
	}

	if user == nil {
		return "", fmt.Errorf("login: incorrect email or password")
	}

	if user.Password == "" {
		return "", fmt.Errorf("login: password not set")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", fmt.Errorf("login: incorrect email or password")
	}

	token, err := Session.Create(user.Id)
	if err != nil {
		slog.Error("login", "err", err)
		return "", fmt.Errorf("login: %w", err)
	}

	return token, nil
}

// return google redirect url
func (user) GoogleSignin() (string, error) {
	clientId, err := Setting.GetGoogleClientID()
	if err != nil {
		return "", err
	}

	siteUrl, err := Setting.GetSiteURL()
	if err != nil {
		return "", err
	}

	params := url.Values{
		"client_id":     {clientId},
		"redirect_uri":  {siteUrl + "/login/google/callback"},
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

// return whether a social login method is configured for a user
func (user) SocialLoginLinked(loginType string, userId int) (bool, error) {
	e, err := db.SocialLoginFind(loginType, userId)
	if err != nil {
		return false, err
	}
	return e != nil, nil
}

// GoogleCallback uses code to obtain access_token from google.
// GoogleCallback returns google user profile
func (user) GoogleCallback(code string) (*OAuthProfile, error) {
	clientId, err := Setting.GetGoogleClientID()
	if err != nil {
		return nil, err
	}

	siteUrl, err := Setting.GetSiteURL()
	if err != nil {
		return nil, err
	}

	secret, err := Setting.GetGoogleSecret()
	if err != nil {
		return nil, err
	}

	type respStruct struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}

	// obtian access token
	form := url.Values{
		"client_id":     {clientId},
		"client_secret": {secret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {siteUrl + "/login/google/callback"},
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

// link a social account to an existing account
func (user) LinkSocialAccount(loginType string, userId int, profile *OAuthProfile) error {
	social, err := db.SocialLoginFind(loginType, userId)
	if err != nil {
		return err
	}

	if social != nil {
		return fmt.Errorf("a %s account is already linked", loginType)
	}

	err = db.SocialLoginCreate(loginType, userId, profile.AccountId)
	return err
}

// signin or register with user profile from social accounts
//
// return a session token
func (user) SigninOrRegisterWithSocial(loginType string, profile *OAuthProfile) (string, error) {
	social, err := db.SocialLoginFindByAccount(loginType, profile.AccountId)
	if err != nil {
		return "", err
	}

	if social == nil {

		// sign up
		randomName := fmt.Sprintf("%s #%d", profile.Name, utils.RandomNumber(100000, 999999))
		userId, err := db.UserCreate(randomName, profile.Email, "", true, RoleUser)
		if err != nil {
			return "", fmt.Errorf("create user: %w", err)
		}

		err = User.LinkSocialAccount(loginType, userId, profile)
		if err != nil {
			return "", fmt.Errorf("link google account: %w", err)
		}

		return Session.Create(userId)

	} else {

		// sign in
		return Session.Create(social.UserId)

	}
}

func (user) UnlinkSocialLogin(loginType string, userId int) error {
	return db.SocialLoginRemove(loginType, userId)
}
