package services

import "img2/db"

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

func (s *setting) GetAll() (map[string]string, error) {
	return db.SettingFindAll()
}

func (s *setting) Set(key, value string) error {
	return db.SetttingUpdate(key, value)
}
