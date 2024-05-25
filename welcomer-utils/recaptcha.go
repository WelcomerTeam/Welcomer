package utils

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

const (
	RecaptchaEndpoint = "https://www.google.com/recaptcha/api/siteverify"
)

type RecaptchaRequest struct {
	Secret    string `json:"secret"`
	Response  string `json:"response"`
	IPAddress string `json:"ip_address"`
}

type RecaptchaResponse struct {
	ChallengeTimestamp time.Time `json:"challenge_ts"`
	Action             string    `json:"action"`
	Hostname           string    `json:"hostname"`
	ErrorCodes         []string  `json:"error-codes"`
	Score              float64   `json:"score"`
	Success            bool      `json:"success"`
}

func ValidateRecaptcha(logger zerolog.Logger, response string, ipAddress string) (float64, error) {
	reqBody := url.Values{}
	reqBody.Set("secret", os.Getenv("RECAPTCHA_SECRET"))
	reqBody.Set("response", response)
	reqBody.Set("ip_address", ipAddress)

	req, err := http.NewRequest(http.MethodPost, RecaptchaEndpoint, strings.NewReader(reqBody.Encode()))
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create reCAPTCHA request")

		return -1, err
	}

	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error().Err(err).Int("status_code", resp.StatusCode).Msg("Failed to send reCAPTCHA request")

		return -1, err
	}

	defer resp.Body.Close()

	var recaptchaResponse RecaptchaResponse

	err = json.NewDecoder(resp.Body).Decode(&recaptchaResponse)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to parse reCAPTCHA response")

		return -1, err
	}

	return recaptchaResponse.Score, nil
}
