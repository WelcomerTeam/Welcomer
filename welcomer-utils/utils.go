package utils

import (
	"encoding/json"
	"fmt"
	"image/color"
	"math/rand"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	urlverifier "github.com/davidmytton/url-verifier"
	"github.com/jackc/pgtype"
	gotils_strconv "github.com/savsgio/gotils/strconv"
)

var (
	True  = true
	False = false
)

var verifier = urlverifier.NewVerifier()

const (
	hexBase      = 16
	int64Base    = 10
	int64BitSize = 64

	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	UserAgent = "WelcomerService (https://github.com/WelcomerTeam/Welcomer)"
)

func Itoa(v int64) string {
	return strconv.FormatInt(v, int64Base)
}

func Atoi(s string) (int64, error) {
	return strconv.ParseInt(s, int64Base, int64BitSize)
}

// RetryWithFallback will attempt to run a function. If the function fails, it will run a fallback function.
// predictate is a function that will determine if the error is retryable and is optional.
// When the fallback function succeeds, the original function will be called again.
func RetryWithFallback(function func() error, fallbackFunction func() error, predictate func(error) bool) (err error) {
	err = function()
	if err != nil && (predictate == nil || predictate(err)) {
		err = fallbackFunction()
		if err == nil {
			return function()
		}
	}

	return
}

func RandStringBytesRmndr(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}

	return string(b)
}

// StringToJsonLiteral converts a string to a json.RawMessage.
func StringToJsonLiteral(s string) json.RawMessage {
	return json.RawMessage([]byte(`"` + s + `"`))
}

func ToPointer[K any](k K) *K {
	return &k
}

func FormatTextStroke(v bool) int {
	if v {
		return 4
	}

	return 0
}

func SetupJSONB(jsonb pgtype.JSONB) pgtype.JSONB {
	if jsonb.Status == pgtype.Undefined || len(jsonb.Bytes) == 0 {
		jsonb.Status = pgtype.Present
		jsonb.Bytes = []byte{123, 125} // {}
	}

	return jsonb
}

func ConvertToRGBA(int32Color int64) color.RGBA {
	alpha := uint8(int32Color >> 24 & 0xFF)
	red := uint8(int32Color >> 16 & 0xFF)
	green := uint8(int32Color >> 8 & 0xFF)
	blue := uint8(int32Color & 0xFF)

	return color.RGBA{R: red, G: green, B: blue, A: alpha}
}

func TryParseBool(str string) bool {
	boolean, _ := strconv.ParseBool(str)

	return boolean
}

func TryParseInt(str string) int64 {
	integer, _ := strconv.ParseInt(str, int64Base, int64BitSize)

	return integer
}

func TryParseFloat(str string) float64 {
	float, _ := strconv.ParseFloat(str, int64BitSize)

	return float
}

func IsValidUnsplashID(str string) bool {
	return unsplashRegex.MatchString(str)
}

func IsValidColour(str string) bool {
	_, err := ParseColour(str, "")

	return err == nil
}

func IsValidInteger(str string) bool {
	_, err := strconv.ParseInt(str, int64Base, int64BitSize)

	return err == nil
}

func IsValidHex(str string, allowAlpha bool) bool {
	if len(str) != 6 && (!allowAlpha || len(str) != 8) {
		return false
	}

	_, err := strconv.ParseUint(str, hexBase, int64BitSize)

	return err == nil
}

func IsValidBackground(s string) bool {
	_, err := ParseBackground(s)

	return err == nil
}

func IsValidEmbed(s string) bool {
	var upe UserProvidedEmbed

	err := json.Unmarshal(gotils_strconv.S2B(s), &upe)

	return err == nil
}

func IsValidImageAlignment(value string) bool {
	_, err := ParseImageAlignment(value)

	return err == nil
}

func IsValidImageTheme(value string) bool {
	_, err := ParseImageTheme(value)

	return err == nil
}

func IsValidImageProfileBorderType(value string) bool {
	_, err := ParseImageProfileBorderType(value)

	return err == nil
}

// Validates a URL and prevents SSRF.
func IsValidURL(url string) (*url.URL, bool) {
	result, err := verifier.Verify(url)
	if err != nil {
		return nil, false
	}

	if !result.IsURL || !result.IsRFC3986URL {
		return nil, false
	}

	if !isValidHostname(result.URLComponents.Hostname()) {
		return nil, false
	}

	return result.URLComponents, true
}

func isValidHostname(host string) bool {
	ips, err := net.LookupIP(host)
	if err != nil {
		return false
	}

	// Check each IP to see if it is an internal IP
	for _, ip := range ips {
		if ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsInterfaceLocalMulticast() || ip.IsUnspecified() {
			return false
		}
	}

	return true
}

// ParseColour parses a colour and returns RGBA.
// Expected formats:
// #FFAAAA
// #FFAAAAFF
// RGBA(255, 255, 255, 0.1)
// RGB(255, 255, 255)
func ParseColour(str string, defaultValue string) (*color.RGBA, error) {
	str = strings.TrimSpace(str)

	if str == "" {
		if defaultValue == "" {
			str = fallbackColour
		} else {
			str = defaultValue
		}
	}

	switch {
	case strings.HasPrefix(str, RGBPrefix):
		if strings.HasPrefix(str, RGBAPrefix) {
			// Validate against RGBA regex
			if RGBARegex.MatchString(str) {
				str = strings.TrimPrefix(str, RGBAPrefix) // (255, 255, 255, 0.1)
				str = strings.TrimPrefix(str, "(")        // 255, 255, 255, 0.1)
				str = strings.TrimSuffix(str, ")")        // 255, 255, 255, 0.1
				values := strings.Split(str, ",")         // ["255", " 255", " 255, " 0.1"]

				var alphaFloat float64
				var err error

				if len(values) >= 4 {
					alphaFloat, err = strconv.ParseFloat(strings.TrimSpace(values[3]), int64BitSize) // 0.1
				} else {
					alphaFloat = 1
				}

				if err == nil && len(values) >= 3 {
					var alphaInt uint8

					// If our float is above one, we will assume alpha max is 255 instead of 1.
					// We store all values as 255 so we multiply the float by 255 if it is not max 1.
					if alphaFloat > 1 {
						alphaInt = uint8(alphaFloat)
					} else {
						alphaInt = uint8(alphaFloat * 255)
					}

					// We can assume these values are ints due to our regex only allowing numerical values.
					colourR, _ := strconv.ParseInt(strings.TrimSpace(values[0]), int64Base, int64BitSize)
					colourG, _ := strconv.ParseInt(strings.TrimSpace(values[1]), int64Base, int64BitSize)
					colourB, _ := strconv.ParseInt(strings.TrimSpace(values[2]), int64Base, int64BitSize)

					if colourR > 255 || colourR < 0 {
						return nil, ErrInvalidColour
					}

					if colourG > 255 || colourG < 0 {
						return nil, ErrInvalidColour
					}

					if colourB > 255 || colourB < 0 {
						return nil, ErrInvalidColour
					}

					return &color.RGBA{uint8(colourR), uint8(colourG), uint8(colourB), alphaInt}, nil
				}
			}
		} else {
			// Validate against RGB regex
			if RGBRegex.MatchString(str) {
				str = strings.TrimPrefix(str, RGBPrefix) // (255, 255, 255)
				str = strings.TrimPrefix(str, "(")       // 255, 255, 255)
				str = strings.TrimSuffix(str, ")")       // 255, 255, 255
				values := strings.Split(str, ",")        // ["255", " 255", " 255]

				if values == nil {
					return nil, ErrInvalidColour
				}

				// We can assume these values are ints due to our regex only allowing numerical values.
				colourR, _ := strconv.ParseInt(strings.TrimSpace(values[0]), int64Base, int64BitSize)
				colourG, _ := strconv.ParseInt(strings.TrimSpace(values[1]), int64Base, int64BitSize)
				colourB, _ := strconv.ParseInt(strings.TrimSpace(values[2]), int64Base, int64BitSize)

				if colourR > 255 || colourR < 0 {
					return nil, ErrInvalidColour
				}

				if colourG > 255 || colourG < 0 {
					return nil, ErrInvalidColour
				}

				if colourB > 255 || colourB < 0 {
					return nil, ErrInvalidColour
				}

				return &color.RGBA{uint8(colourR), uint8(colourG), uint8(colourB), 255}, nil
			}
		}
	default:
		str = strings.TrimPrefix(str, "#")
		if IsValidHex(str, true) {
			// We can assume these values are ints due to isValidHex.
			colourR, _ := strconv.ParseInt(strings.TrimSpace(str[0:2]), hexBase, int64BitSize)
			colourG, _ := strconv.ParseInt(strings.TrimSpace(str[2:4]), hexBase, int64BitSize)
			colourB, _ := strconv.ParseInt(strings.TrimSpace(str[4:6]), hexBase, int64BitSize)

			var colourA int64

			if len(str) == 8 {
				colourA, _ = strconv.ParseInt(strings.TrimSpace(str[6:8]), hexBase, int64BitSize)
			} else {
				colourA = 255
			}

			if colourR > 255 || colourR < 0 {
				return nil, ErrInvalidColour
			}

			if colourG > 255 || colourG < 0 {
				return nil, ErrInvalidColour
			}

			if colourB > 255 || colourB < 0 {
				return nil, ErrInvalidColour
			}

			return &color.RGBA{uint8(colourR), uint8(colourG), uint8(colourB), uint8(colourA)}, nil
		}
	}

	return nil, ErrInvalidColour
}

// IsJSONBEmpty checks if a byte slice is empty or is a JSON empty object.
func IsJSONBEmpty(b []byte) bool {
	return len(b) == 0 || (len(b) == 2 && b[0] == '{' && b[1] == '}')
}

// IsMessageParamsEmpty checks if the given message parameters are empty.
// It returns true if the content and all the fields in the embeds are empty, otherwise it returns false.
func IsMessageParamsEmpty(m discord.MessageParams) bool {
	if m.Content != "" {
		return false
	}

	if len(m.Files) > 0 {
		return false
	}

	if len(m.Embeds) == 0 {
		return true
	}

	for _, embed := range m.Embeds {
		if embed.Title != "" || embed.Description != "" || embed.URL != "" || len(embed.Fields) > 0 || embed.Color != 0 {
			return false
		}

		// Check each field in the embed
		for _, field := range embed.Fields {
			if field.Name != "" || field.Value != "" || field.Inline {
				return false
			}
		}
	}

	return true
}

func HumanizeDuration(seconds int) string {
	duration := time.Duration(seconds) * time.Second
	years := int(duration.Hours() / 24 / 365)
	days := int(duration.Hours()/24) % 365
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	var result strings.Builder

	if years > 0 {
		result.WriteString(fmt.Sprintf("%d year", years))
		if years > 1 {
			result.WriteString("s")
		}
	}

	if days > 0 {
		if result.Len() > 0 {
			result.WriteString(", ")
		}
		result.WriteString(fmt.Sprintf("%d day", days))
		if days > 1 {
			result.WriteString("s")
		}
	}

	if hours > 0 {
		if result.Len() > 0 {
			result.WriteString(", ")
		}
		result.WriteString(fmt.Sprintf("%d hour", hours))
		if hours > 1 {
			result.WriteString("s")
		}
	}

	if minutes > 0 {
		if result.Len() > 0 {
			result.WriteString(" and ")
		}
		result.WriteString(fmt.Sprintf("%d minute", minutes))
		if minutes > 1 {
			result.WriteString("s")
		}
	}

	return result.String()
}

func If[T any](condition bool, trueValue, falseValue T) T {
	if condition {
		return trueValue
	}

	return falseValue
}

func SliceContains[T comparable](slice []T, value T) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}

	return false
}
