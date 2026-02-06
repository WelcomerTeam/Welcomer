package welcomer

func GuildHasFeature(guildFeatures []GuildFeature, feature GuildFeature) bool {
	for _, gf := range guildFeatures {
		if gf == feature {
			return true
		}
	}

	return false
}
