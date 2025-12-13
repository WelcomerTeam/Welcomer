package welcomer

type GuildFeature string

// GuildFeatures are features that can be enabled or disabled per guild.
// These are not membership tied, but rather guild specific features and
// are used to gate access to certain features or to limit a beta test.

const (
	GuildFeatureCustomWelcomerImageBuilderBeta GuildFeature = "custom_welcomer_image_builder_beta"
)

// List of guild features that can be opted into via the /optin command.

var optinGuildFeatures = map[GuildFeature]bool{
	GuildFeatureCustomWelcomerImageBuilderBeta: true,
}
