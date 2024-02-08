package services

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
)

type captcha struct{}

var CAPTCHA = captcha{}

func (*captcha) VerifyReCAPTCHA(response string) bool {

	secret, err := Setting.GetReCaptchaServer()
	if err != nil {
		slog.Error("verify recaptcha", "err", err)
		return false
	}

	form := url.Values{}
	form.Set("response", response)
	form.Set("secret", secret)

	resp, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify", form)
	if err != nil {
		slog.Error("verify recaptcha", "err", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("verify recaptcha", "status", resp.Status)
		return false
	}

	var data struct {
		Success    bool     `json:"success"`
		ErrorCodes []string `json:"error-codes"`
	}

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		slog.Error("verify recaptcha", "err", fmt.Errorf("malformated json: %w", err))
		return false
	}

	if data.Success {
		return true
	}

	slog.Debug("verify recaptcha", "error-codes", fmt.Sprintf("%v", data.ErrorCodes))
	return false
}

func (*captcha) VerifyHCaptcha(response string) bool {
	secret, err := Setting.GetHCaptchaServer()
	if err != nil {
		slog.Error("verify hcaptcha", "err", err)
		return false
	}

	form := url.Values{}
	form.Set("response", response)
	form.Set("secret", secret)

	resp, err := http.PostForm("https://api.hcaptcha.com/siteverify", form)
	if err != nil {
		slog.Error("verify hcaptcha", "err", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("verify recaptcha", "status", resp.Status)
		return false
	}

	var data struct {
		Success    bool     `json:"success"`
		ErrorCodes []string `json:"error-codes"`
	}

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		slog.Error("verify hcaptcha", "err", fmt.Errorf("malformated json: %w", err))
		return false
	}

	if data.Success {
		return true
	}

	slog.Debug("verify hcaptcha", "error-codes", fmt.Sprintf("%v", data.ErrorCodes))
	return false
}
