package utils

import (
	"github.com/ua-parser/uap-go/uaparser"
)

var userAgentParser, _ = uaparser.NewFromBytes(uaparser.DefinitionYaml)

func ParseUserAgent(userAgent string) (familyName, familyVersion, osName, osVersion string) {
	client := userAgentParser.Parse(userAgent)

	return client.UserAgent.Family, client.UserAgent.ToVersionString(), client.Os.Family, client.Os.ToVersionString()
}
