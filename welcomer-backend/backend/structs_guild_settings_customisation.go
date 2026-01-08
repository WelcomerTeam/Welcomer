package backend

type GuildSettingsCustomisation struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Banner   string `json:"banner"`
	Bio      string `json:"bio"`
}
