package welcomer

import (
	"encoding/json"

	"github.com/WelcomerTeam/Discord/discord"
)

func UnmarshalRolesListJSON(rolesJSON []byte) (roles []discord.Snowflake) {
	_ = json.Unmarshal(rolesJSON, &roles)

	return
}

func MarshalRolesListJSON(roles []discord.Snowflake) (rolesJSON []byte) {
	rolesJSON, _ = json.Marshal(roles)

	return
}

type GiveawayPrize struct {
	Title string `json:"title"`
	Count int    `json:"count"`
}

func UnmarshalGiveawayPrizeJSON(prizeJSON []byte) (prize []GiveawayPrize) {
	_ = json.Unmarshal(prizeJSON, &prize)

	// Ensure count is at least 1.
	for i := range prize {
		if prize[i].Count <= 0 {
			prize[i].Count = 1
		}
	}

	return
}

func MarshalGiveawayPrizeJSON(prize []GiveawayPrize) (prizeJSON []byte) {
	prizeJSON, _ = json.Marshal(prize)

	return
}
