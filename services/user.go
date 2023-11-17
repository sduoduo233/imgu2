package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"img2/db"
	"img2/services/emails"
	"img2/utils"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
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

// used for signing email verification & password reset links
var jwtSecret = ""

type user struct{}

var User = user{}

// user profile obtained from oauth services (e.g. google user profile)
type OAuthProfile struct {
	Name      string
	Email     string
	AccountId string
}

func getJWTSecret() string {
	if jwtSecret == "" {
		jwtSecret = os.Getenv("IMG2_JWT_SECRET")
		if jwtSecret == "" {
			panic("IMG2_JWT_SECRET should not be empty")
		}
	}

	return jwtSecret
}

func (user) FindByEmail(email string) (*db.User, error) {
	return db.UserFindByEmail(email)
}

func (user) FindById(userId int) (*db.User, error) {
	return db.UserFindById(userId)
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

// return github oauth redirect url
func (user) GithubSignin() (string, error) {
	clientId, err := Setting.GetGithubClientID()
	if err != nil {
		return "", err
	}

	params := url.Values{
		"client_id": {clientId},
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

// obtain access token using code
// return github user profile
func (user) GithubCallback(code string) (*OAuthProfile, error) {
	clientId, err := Setting.GetGithubClientID()
	if err != nil {
		return nil, err
	}

	secret, err := Setting.GetGithubSecret()
	if err != nil {
		return nil, err
	}

	// obtain access token
	type reqStruct struct {
		ClientId     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		Code         string `json:"code"`
	}
	body, err := json.Marshal(&reqStruct{
		ClientId:     clientId,
		ClientSecret: secret,
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
			profile.Email = email.Email
			return &profile, nil
		}
	}

	return nil, fmt.Errorf("nil")
}

func (user) ChangeUsername(userId int, username string) error {
	return db.UserChangeUsername(userId, username)
}

// update email and set email verified to false
func (user) ChangeEmail(userId int, email string) error {
	return db.UserChangeEmail(userId, email)
}

// send verification email
func (user) SendVerificationEmail(userId int) error {
	u, err := User.FindById(userId)
	if err != nil {
		return err
	}

	if u == nil {
		return fmt.Errorf("user not found: %d", userId)
	}

	if u.EmailVerified {
		return fmt.Errorf("email already verified: %d", userId)
	}

	siteName, err := Setting.GetSiteName()
	if err != nil {
		return err
	}

	siteUrl, err := Setting.GetSiteURL()
	if err != nil {
		return err
	}

	// generate verification url
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Minute * 30).Unix(), // Expiration Time
		"nbf": time.Now().Unix(),                       // Not Before
		"sub": u.Email,
		"aud": "email_verification",
	})

	signedToken, err := token.SignedString([]byte(getJWTSecret()))
	if err != nil {
		return fmt.Errorf("jwt sign: %w", err)
	}

	// generate email content
	tpl, err := template.New("verification").Parse(emails.VERIFICATION)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	buf := new(bytes.Buffer)

	err = tpl.Execute(buf, map[string]string{
		"username": u.Username,
		"name":     siteName,
		"link":     siteUrl + "/verify-email?token=" + signedToken,
		"email":    u.Email,
	})
	if err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	// send email
	err = Mailer.SendMail(u.Email, "Confirm your "+siteName+" account", buf.String())
	if err != nil {
		return err
	}

	return nil
}

// verify email stated in the jwt token
func (user) VerifyEmail(token string) error {
	// parse and validate jwt token
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(getJWTSecret()), nil
	})

	if err != nil {
		return fmt.Errorf("parse token: %w", err)
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("parse token: Claims is not jwt.MapClaims")
	}

	if claims["aud"] != "email_verification" {
		return fmt.Errorf("invalid aud: %v", claims["aud"])
	}

	email, ok := claims["sub"].(string)
	if !ok {
		return fmt.Errorf("invalid email: %v", claims["sub"])
	}

	// mark email as verified
	err = db.UserVerifyEmail(email)
	return err
}
