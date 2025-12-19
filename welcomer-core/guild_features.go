package welcomer

// GuildFeatures are features that can be enabled or disabled per guild.
// These are not membership tied, but rather guild specific features and
// are used to gate access to certain features or to limit a beta test.

//go:generate go-enum -f=$GOFILE --marshal

// ENUM(CustomWelcomerImageBuilder)
type GuildFeature string

// List of guild features that can be opted into via the /optin command.

var OptinGuildFeatures = map[GuildFeature]bool{
	GuildFeatureCustomWelcomerImageBuilder: true,
}
