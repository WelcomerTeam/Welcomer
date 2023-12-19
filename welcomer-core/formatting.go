package welcomer

import (
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/WelcomerTeam/Discord/discord"
	mustache "github.com/WelcomerTeam/Mustachvulate"
	sandwich "github.com/WelcomerTeam/Sandwich/sandwich"
)

func GatherFunctions() (funcs map[string]govaluate.ExpressionFunction) {
	return
}

func GatherVariables(eventCtx *sandwich.EventContext, member discord.GuildMember, guild discord.Guild) (vars map[string]interface{}) {
	vars = make(map[string]interface{})

	vars["User"] = StubUser{
		ID:            member.User.ID,
		Name:          getUserDisplayName(member.User),
		Username:      member.User.Username,
		Discriminator: member.User.Discriminator,
		GlobalName:    member.User.GlobalName,
		CreatedAt:     StubTime(member.User.ID.Time()),
		JoinedAt:      StubTime(member.JoinedAt),
		Avatar:        getUserAvatar(member.User),
		Bot:           member.User.Bot,
		Pending:       member.Pending,
	}

	vars["Guild"] = StubGuild{
		ID:      guild.ID,
		Name:    guild.Name,
		Icon:    getGuildIcon(guild),
		Splash:  getGuildSplash(guild),
		Members: guild.MemberCount,
		Banner:  getGuildBanner(guild),
	}

	return
}

func FormatString(funcs map[string]govaluate.ExpressionFunction, vars map[string]interface{}, message string) (string, error) {
	tmpl, err := mustache.ParseString(message)
	if err != nil {
		return "", fmt.Errorf("failed to parse string: %v", err)
	}

	out, err := tmpl.Render(funcs, vars)
	if err != nil {
		return "", fmt.Errorf("failed to format string: %v", err)
	}

	return html.UnescapeString(out), nil
}

func getGuildIcon(guild discord.Guild) string {
	if guild.Icon == "" {
		return ""
	}

	if strings.HasPrefix(guild.Icon, "a_") {
		return discord.EndpointCDN + "/" + discord.EndpointGuildIconAnimated(guild.ID.String(), guild.Icon)
	}

	return discord.EndpointCDN + "/" + discord.EndpointGuildIcon(guild.ID.String(), guild.Icon)
}

func getGuildSplash(guild discord.Guild) string {
	if guild.Splash == "" {
		return ""
	}

	return discord.EndpointCDN + "/" + discord.EndpointGuildSplash(guild.ID.String(), guild.Splash)
}

func getGuildBanner(guild discord.Guild) string {
	if guild.Banner == "" {
		return ""
	}

	return discord.EndpointCDN + "/" + discord.EndpointGuildBanner(guild.ID.String(), guild.Banner)
}

func getUserAvatar(user *discord.User) string {
	if user.Avatar == "" {
		if user.Discriminator == "" {
			// If a user is on the new username system, the index is (user_id >> 22) % 6
			return discord.EndpointCDN + "/" + discord.EndpointDefaultUserAvatar(strconv.FormatInt((int64(user.ID)>>22)%6, int64Base))
		}

		// If a user is on the old username system, the index is discriminator % 5
		discriminator, err := strconv.ParseInt(user.Discriminator, int64Base, int64BitSize)
		if err != nil {
			discriminator = 0
		}

		return discord.EndpointCDN + "/" + discord.EndpointDefaultUserAvatar(strconv.FormatInt(discriminator%5, int64Base))
	}

	if strings.HasPrefix(user.Avatar, "a_") {
		// If the avatar has the prefix a_, it is animated.
		return discord.EndpointCDN + "/" + discord.EndpointUserAvatarAnimated(user.ID.String(), user.Avatar)
	}

	return discord.EndpointCDN + "/" + discord.EndpointUserAvatar(user.ID.String(), user.Avatar)
}

func getUserDisplayName(user *discord.User) string {
	if user.GlobalName != "" {
		return user.GlobalName
	}

	if user.Discriminator != "" && user.Discriminator != "0" {
		return user.Username + "#" + user.Discriminator
	}

	return user.Username
}

type StubUser struct {
	ID   discord.Snowflake `json:"id"`
	Name string            `json:"name"`

	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	GlobalName    string `json:"global_name"`

	CreatedAt StubTime `json:"created_at"`
	JoinedAt  StubTime `json:"joined_at"`

	Avatar  string `json:"avatar"`
	Bot     bool   `json:"bot"`
	Pending bool   `json:"pending"`
}

// Guild represents a guild on discord.
type StubGuild struct {
	ID      discord.Snowflake `json:"id"`
	Name    string            `json:"name"`
	Icon    string            `json:"icon"`
	Splash  string            `json:"splash"`
	Members int32             `json:"members"`
	Banner  string            `json:"banner"`
}

type StubTime time.Time

func (s StubTime) String() string {
	return s.Relative()
}

func (s StubTime) Relative() string {
	return "<t:" + strconv.FormatInt(time.Time(s).Unix(), int64Base) + ":R>"
}
