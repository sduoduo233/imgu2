package services

import (
	"fmt"
	"imgu2/db"
	"strconv"
)

type setting struct{}

var Setting = setting{}

const (
	CAPTCHA_NONE      = "none"
	CAPTCHA_RECAPTCHA = "recaptcha"
	CAPTCHA_HCAPTCHA  = "hcaptcha"
)

func (s *setting) GetSiteName() (string, error) {
	return db.SettingFind("SITE_NAME")
}

func (s *setting) GetSiteURL() (string, error) {
	return db.SettingFind("SITE_URL")
}

func (s *setting) GetCAPTCHA() (string, error) {
	return db.SettingFind("CAPTCHA")
}

func (s *setting) GetReCaptchaClient() (string, error) {
	return db.SettingFind("RECAPTCHA_CLIENT")
}

func (s *setting) GetReCaptchaServer() (string, error) {
	return db.SettingFind("RECAPTCHA_SERVER")
}

func (s *setting) GetHCaptchaClient() (string, error) {
	return db.SettingFind("HCAPTCHA_CLIENT")
}

func (s *setting) GetHCaptchaServer() (string, error) {
	return db.SettingFind("HCAPTCHA_SERVER")
}

func (s *setting) GetGoogleClientID() (string, error) {
	return db.SettingFind("GOOGLE_CLIENT_ID")
}

func (s *setting) GetGoogleSecret() (string, error) {
	return db.SettingFind("GOOGLE_SECRET")
}

func (*setting) GetGoogleLogin() (bool, error) {
	s, err := db.SettingFind("GOOGLE_SIGNIN")
	return s == "true", err
}

func (s *setting) GetGithubClientID() (string, error) {
	return db.SettingFind("GITHUB_CLIENT_ID")
}

func (s *setting) GetGithubSecret() (string, error) {
	return db.SettingFind("GITHUB_SECRET")
}

func (*setting) GetGithubLogin() (bool, error) {
	s, err := db.SettingFind("GITHUB_SIGNIN")
	return s == "true", err
}

func (s *setting) GetAll() (map[string]string, error) {
	return db.SettingFindAll()
}

func (s *setting) Set(key, value string) error {
	return db.SetttingUpdate(key, value)
}

func (s *setting) GetLanguage() (string, error) {
	return db.SettingFind("LANGUAGE")
}

func (*setting) GetAllowRegister() (bool, error) {
	s, err := db.SettingFind("ALLOW_REGISTER")
	if err != nil {
		return false, err
	}

	return s == "true", nil
}

func (*setting) IsAVIFEncodingEnabled() (bool, error) {
	s, err := db.SettingFind("AVIF_ENCODING")
	if err != nil {
		return false, err
	}

	return s == "true", nil
}

func (*setting) IsWEBPEncodingEnabled() (bool, error) {
	s, err := db.SettingFind("WEBP_ENCODING")
	if err != nil {
		return false, err
	}

	return s == "true", nil
}

func (*setting) DefaultGroupGuest() (int, error) {
	s, err := db.SettingFind("DEFAULT_GROUP_GUEST")
	if err != nil {
		return 0, err
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("settings: group id is not an integer: %s", s)
	}

	return i, nil
}

func (*setting) DefaultGroupRegistered() (int, error) {
	s, err := db.SettingFind("DEFAULT_GROUP_REGISTERED")
	if err != nil {
		return 0, err
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("settings: group id is not an integer: %s", s)
	}

	return i, nil
}
