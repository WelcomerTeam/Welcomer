package welcomer

import (
	"image/color"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	urlverifier "github.com/davidmytton/url-verifier"
	jsoniter "github.com/json-iterator/go"
)

const (
	hexBase      = 16
	int64Base    = 10
	int64BitSize = 64
)

var verifier = urlverifier.NewVerifier()

func CheckGuildMemberships(memberships []*database.GetUserMembershipsByGuildIDRow) (hasWelcomerPro bool, hasCustomBackgrounds bool) {
	for _, membership := range memberships {
		switch database.MembershipType(membership.MembershipType) {
		case database.MembershipTypeLegacyCustomBackgrounds,
			database.MembershipTypeCustomBackgrounds:
			hasCustomBackgrounds = true
		case database.MembershipTypeLegacyWelcomerPro1,
			database.MembershipTypeLegacyWelcomerPro3,
			database.MembershipTypeLegacyWelcomerPro5,
			database.MembershipTypeWelcomerPro:
			hasWelcomerPro = true
		}
	}

	return
}

func FormatTextStroke(v bool) int {
	if v {
		return 4
	}

	return 0
}

func ConvertToRGBA(int32Color int64) color.RGBA {
	alpha := uint8(int32Color >> 24 & 0xFF)
	red := uint8(int32Color >> 16 & 0xFF)
	green := uint8(int32Color >> 8 & 0xFF)
	blue := uint8(int32Color & 0xFF)

	return color.RGBA{R: red, G: green, B: blue, A: alpha}
}

func MustParseBool(str string) bool {
	boolean, _ := strconv.ParseBool(str)

	return boolean
}

func MustParseInt(str string) int {
	integer, _ := strconv.ParseInt(str, int64Base, int64BitSize)

	return int(integer)
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

	err := jsoniter.UnmarshalFromString(s, &upe)

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

				alphaString := strings.TrimSpace(values[3]) // 0.1
				alphaFloat, err := strconv.ParseFloat(alphaString, int64BitSize)

				if err == nil {
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
