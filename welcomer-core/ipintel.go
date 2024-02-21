package welcomer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/rs/zerolog"
)

const (
	IPIntelEndpoint = "https://check.getipintel.net/check.php"
)

type IPIntelResponse struct {
	Success      string  `json:"success"`
	ResultString string  `json:"result"`
	Result       float64 `json:"-"`
}

var errorCodes = map[float64]string{
	-1: "Invalid no input",
	-2: "Invalid IP address",
	-3: "Unroutable address / private address",
	-4: "Unable to reach database",
	-5: "Your connecting IP has been banned from the system or you do not have permission to access a particular service.",
	-6: "You did not provide any contact information with your query or the contact information is invalid.",
}

func CheckIPIntel(logger zerolog.Logger, ipAddress string) (response float64, err error) {
	reqParams := url.Values{}
	reqParams.Set("ip", ipAddress)
	reqParams.Set("contact", os.Getenv("IPINTEL_CONTACT"))
	reqParams.Set("format", "json")
	reqParams.Set("oflags", "b")

	req, err := http.NewRequest(http.MethodGet, IPIntelEndpoint+"?"+reqParams.Encode(), nil)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create IPIntel request")

		return -1, err
	}

	req.Header.Set("User-Agent", UserAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error().Err(err).Int("status_code", resp.StatusCode).Msg("Failed to send IPIntel request")

		return -1, err
	}

	defer resp.Body.Close()

	var ipIntelResponse IPIntelResponse

	err = json.NewDecoder(resp.Body).Decode(&ipIntelResponse)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to decode IPIntel response")

		return -1, err
	}

	ipIntelResponse.Result, err = strconv.ParseFloat(ipIntelResponse.ResultString, 64)
	if err != nil {
		logger.Error().Err(err).Str("response", ipIntelResponse.ResultString).Msg("Failed to parse IPIntel response result")

		return -1, err
	}

	if ipIntelResponse.Result < 0 {
		logger.Error().Float64("result", ipIntelResponse.Result).Msg("IPIntel returned an error")

		if message, ok := errorCodes[ipIntelResponse.Result]; ok {
			return -1, fmt.Errorf(fmt.Sprintf("ipintel failed with code %f: %s", ipIntelResponse.Result, message))
		}

		return -1, fmt.Errorf(fmt.Sprintf("ipintel failed with code %f: unknown error", ipIntelResponse.Result))
	}

	return ipIntelResponse.Result, nil
}
