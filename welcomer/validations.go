package welcomer

import (
	"strconv"

	jsoniter "github.com/json-iterator/go"
)

const (
	hexBase      = 16
	int64Base    = 10
	int64BitSize = 64
)

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
