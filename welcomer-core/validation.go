package welcomer

import "regexp"

const (
	MaxRuleCount  = 25
	MaxRuleLength = 250
)

var regexDiscordToken = regexp.MustCompile(`^[A-Za-z0-9_\-]{24,28}\.[A-Za-z0-9_\-]{6}\.[A-Za-z0-9_\-]{27,38}$`)

func IsValidPublicKey(publicKey string) bool {
	// Public key should be a 64-character hexadecimal string

	if len(publicKey) != 64 {
		return false
	}

	for _, char := range publicKey {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') && (char < 'A' || char > 'F') {
			return false
		}
	}

	return true
}

func IsValidDiscordToken(token string) bool {
	return regexDiscordToken.MatchString(token)
}
