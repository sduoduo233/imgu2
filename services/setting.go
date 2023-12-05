package services

import (
	"fmt"
	"img2/db"
	"strconv"
)

type setting struct{}

var Setting = setting{}

const (
	CAPTCHA_NONE = "none"
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

func (*setting) GetMaxImageSize() (uint, error) {
	s, err := db.SettingFind("MAX_IMAGE_SIZE")
	if err != nil {
		return 0, err
	}

	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("strconv: %w", err)
	}

	if n < 0 {
		return 0, fmt.Errorf("negative MAX_IMAGE_SIZE")
	}

	return uint(n), nil
}
