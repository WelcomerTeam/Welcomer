package welcomer

func GuildHasFeature(guildFeatures []GuildFeature, feature GuildFeature) bool {
	// temp
	return True

	for _, gf := range guildFeatures {
		if gf == feature {
			return true
		}
	}

	return false
}
