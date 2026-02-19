package welcomer

import "slices"

func GuildHasFeature(guildFeatures []GuildFeature, feature GuildFeature) bool {
	return slices.Contains(guildFeatures, feature)
}
