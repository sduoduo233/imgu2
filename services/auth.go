package services

import (
	"fmt"
	"imgu2/db"
	"imgu2/services/oauth"
	"log/slog"
	"os"

	"golang.org/x/crypto/bcrypt"
)

type auth struct {
	googleOAuth oauth.OAuth
	githubOAuth oauth.OAuth
}

var Auth = auth{}

// used for signing email verification & password reset links
var jwtSecret = ""

func getJWTSecret() string {
	if jwtSecret == "" {
		jwtSecret = os.Getenv("IMGU2_JWT_SECRET")
		if jwtSecret == "" {
			panic("IMGU2_JWT_SECRET should not be empty")
		}
	}

	return jwtSecret
}

// login the user using email and password
//
// return a new token
func (*auth) Login(email string, password string) (string, error) {
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

// return whether a social login method is configured for a user
func (*auth) SocialLoginLinked(loginType string, userId int) (bool, error) {
	e, err := db.SocialLoginFind(loginType, userId)
	if err != nil {
		return false, err
	}
	return e != nil, nil
}

// link a social account to an existing account
func (auth) LinkSocialAccount(loginType string, userId int, profile *oauth.OAuthProfile) error {
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
func (*auth) SigninOrRegisterWithSocial(loginType string, profile *oauth.OAuthProfile) (string, error) {
	social, err := db.SocialLoginFindByAccount(loginType, profile.AccountId)
	if err != nil {
		return "", err
	}

	if social == nil {

		allowRegister, err := Setting.GetAllowRegister()
		if err != nil {
			return "", err
		}

		if !allowRegister {
			return "", fmt.Errorf("registration is disabled")
		}

		groupId, err := Setting.DefaultGroupRegistered()
		if err != nil {
			return "", err
		}

		// sign up
		randomName := fmt.Sprintf("%s #%d", profile.Name, RandomNumber(100000, 999999))
		userId, err := db.UserCreate(randomName, profile.Email, "", true, RoleUser, groupId)
		if err != nil {
			return "", fmt.Errorf("create user: %w", err)
		}

		err = Auth.LinkSocialAccount(loginType, userId, profile)
		if err != nil {
			return "", fmt.Errorf("link google account: %w", err)
		}

		return Session.Create(userId)

	} else {

		// sign in
		return Session.Create(social.UserId)

	}
}

func (*auth) UnlinkSocialLogin(loginType string, userId int) error {
	return db.SocialLoginRemove(loginType, userId)
}

// initialize oauth providers
func (a *auth) InitOAuthProviders() error {
	googleOAuth, err := Setting.GetGoogleLogin()
	if err != nil {
		return err
	}

	if googleOAuth {
		clientId, err := Setting.GetGoogleClientID()
		if err != nil {
			return err
		}

		secret, err := Setting.GetGoogleSecret()
		if err != nil {
			return err
		}

		siteUrl, err := Setting.GetSiteURL()
		if err != nil {
			return err
		}

		a.googleOAuth = oauth.NewGoogleOAuth(clientId, secret, siteUrl)
	}

	githubOAuth, err := Setting.GetGithubLogin()
	if err != nil {
		return err
	}

	if githubOAuth {
		clientId, err := Setting.GetGithubClientID()
		if err != nil {
			return err
		}

		secret, err := Setting.GetGithubSecret()
		if err != nil {
			return err
		}

		a.githubOAuth = oauth.NewGithubOAuth(clientId, secret)
	}

	return nil
}

func (a *auth) GoogleOAuth() oauth.OAuth {
	return a.googleOAuth
}

func (a *auth) GithubOAuth() oauth.OAuth {
	return a.githubOAuth
}
