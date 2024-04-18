package services

import (
	"bytes"
	"fmt"
	"html/template"
	"imgu2/db"
	"imgu2/services/emails"
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

type user struct{}

var User = user{}

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

func (user) ChangeUsername(userId int, username string) error {
	return db.UserChangeUsername(userId, username)
}

// send a verification link to the new email address
//
// email address is only changed after clicking the verification link
func (user) ChangeEmail(userId int, newEmail string) error {

	u, err := User.FindById(userId)
	if err != nil {
		return err
	}

	if u == nil {
		return fmt.Errorf("user not found: %d", userId)
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
		"exp":       time.Now().Add(time.Minute * 30).Unix(), // Expiration Time
		"nbf":       time.Now().Unix(),                       // Not Before
		"sub":       u.Id,
		"new_email": newEmail,
		"aud":       "email_change",
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
		"link":     siteUrl + "/callback/verify-email-change?token=" + signedToken,
		"email":    newEmail,
	})
	if err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	// send email
	err = Mailer.SendMail(newEmail, "Confirm your "+siteName+" account", buf.String())
	if err != nil {
		return err
	}

	return nil
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
		"link":     siteUrl + "/callback/verify-email?token=" + signedToken,
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

// change email address
func (user) ChangeEmailCallback(token string) error {
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

	if claims["aud"] != "email_change" {
		return fmt.Errorf("invalid aud: %v", claims["aud"])
	}

	userId, ok := claims["sub"].(float64) // golang uses float64 for JSON numbers
	if !ok {
		return fmt.Errorf("invalid user id: %v", claims["sub"])
	}

	email, ok := claims["new_email"].(string)
	if !ok {
		return fmt.Errorf("invalid email: %v", claims["new_email"])
	}

	// update email address and mark as verified
	err = db.UserChangeEmail(int(userId), email)
	return err
}

// register a new account
//
// return a session token for the new account
func (*user) Register(username, email, password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 0)
	if err != nil {
		return "", fmt.Errorf("bcrypt: %w", err)
	}

	groupId, err := Setting.DefaultGroupRegistered()
	if err != nil {
		return "", err
	}

	id, err := db.UserCreate(username, email, string(hashed), false, RoleUser, groupId)
	if err != nil {
		return "", fmt.Errorf("create user: %w", err)
	}

	token, err := Session.Create(id)
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}

	return token, nil
}

// ResetPassword sends a password reset email to user
//
// ResetPassword does not return error if email is not found
func (*user) ResetPassword(email string) error {
	u, err := User.FindByEmail(email)
	if err != nil {
		return err
	}

	if u == nil {
		// ignore if user not found
		return nil
	}

	if !u.EmailVerified {
		return fmt.Errorf("email unverified")
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
		"sub": u.Id,
		"aud": "reset_password",
	})

	signedToken, err := token.SignedString([]byte(getJWTSecret()))
	if err != nil {
		return fmt.Errorf("jwt sign: %w", err)
	}

	// generate email content
	tpl, err := template.New("reset_password").Parse(emails.RESET_PASSWORD)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	buf := new(bytes.Buffer)

	err = tpl.Execute(buf, map[string]string{
		"username": u.Username,
		"name":     siteName,
		"link":     siteUrl + "/callback/reset-password?token=" + signedToken,
	})
	if err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	// send email
	err = Mailer.SendMail(u.Email, "Reset password for your "+siteName+" account", buf.String())
	if err != nil {
		return err
	}

	return nil
}

// ResetPasswordCallback verifies the JWT token and changes user's password
func (*user) ResetPasswordCallback(token string, password string) error {
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

	if claims["aud"] != "reset_password" {
		return fmt.Errorf("invalid aud: %v", claims["aud"])
	}

	userId, ok := claims["sub"].(float64) // golang uses float64 for JSON numbers
	if !ok {
		return fmt.Errorf("invalid user id: %v", claims["sub"])
	}

	// change password
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 0)
	if err != nil {
		return fmt.Errorf("bcrypt: %w", err)
	}
	err = db.UserChangePassword(int(userId), string(hashed))
	if err != nil {
		return err
	}

	return nil
}

func (*user) FindAll(page int) ([]db.User, error) {
	const pageSize = 100
	return db.UserFindAll(page*pageSize, pageSize)
}

func (*user) CountAll() (int, error) {
	return db.UserCount()
}

func (*user) ChangeRole(id int, role int) error {
	if role != RoleAdmin && role != RoleUser && role != RoleBanned {
		return fmt.Errorf("unknown role: %d", role)
	}
	return db.UserChangeRole(id, role)
}

func (*user) ChangeGroup(id int, group int) error {
	return db.UserChangeGroup(id, group)
}

// Set the time when the user group expires.
func (*user) ChangeGroupExpire(id int, expire int) error {
	return db.UserChangeGroupExpire(id, expire)
}
