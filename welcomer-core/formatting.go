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

// ordinal takes 1 argument but 0 was given
// ordinal argument 1 type "string" is not supported

func AssertLength(name string, expectedLength int, arguments ...interface{}) (err error) {
	if len(arguments) != expectedLength {
		return fmt.Errorf("%s takes %d argument(s) but %d was given", name, expectedLength, len(arguments))
	}

	return nil
}

func GatherFunctions() (funcs map[string]govaluate.ExpressionFunction) {
	funcs = map[string]govaluate.ExpressionFunction{}

	funcs["Ordinal"] = func(arguments ...interface{}) (interface{}, error) {
		if err := AssertLength("Ordinal", 1, arguments...); err != nil {
			return nil, err
		}

		argument, ok := arguments[0].(float64)
		if !ok {
			return nil, fmt.Errorf("Ordinal argument 1 is not supported")
		}

		var suffix string

		switch int64(argument) % 10 {
		case 1:
			suffix = "st"
		case 2:
			suffix = "nd"
		case 3:
			suffix = "rd"
		default:
			suffix = "th"
		}

		return Itoa(int64(argument)) + suffix, nil
	}

	return
}

func GatherVariables(eventCtx *sandwich.EventContext, member discord.GuildMember, guild discord.Guild) (vars map[string]interface{}) {
	vars = make(map[string]interface{})

	vars["User"] = StubUser{
		ID:            member.User.ID,
		Name:          GetUserDisplayName(member.User),
		Username:      member.User.Username,
		Discriminator: member.User.Discriminator,
		GlobalName:    member.User.GlobalName,
		Mention:       "<@" + member.User.ID.String() + ">",
		CreatedAt:     StubTime(member.User.ID.Time()),
		JoinedAt:      StubTime(member.JoinedAt),
		Avatar:        GetUserAvatar(member.User),
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

func GetUserAvatar(user *discord.User) string {
	if user.Avatar == "" {
		if user.Discriminator == "" {
			// If a user is on the new username system, the index is (user_id >> 22) % 6
			return discord.EndpointCDN + "/" + discord.EndpointDefaultUserAvatar(Itoa((int64(user.ID)>>22)%6))
		}

		// If a user is on the old username system, the index is discriminator % 5
		discriminator, err := strconv.ParseInt(user.Discriminator, int64Base, int64BitSize)
		if err != nil {
			discriminator = 0
		}

		return discord.EndpointCDN + "/" + discord.EndpointDefaultUserAvatar(Itoa(discriminator%5))
	}

	if strings.HasPrefix(user.Avatar, "a_") {
		// If the avatar has the prefix a_, it is animated.
		return discord.EndpointCDN + "/" + discord.EndpointUserAvatarAnimated(user.ID.String(), user.Avatar)
	}

	return discord.EndpointCDN + "/" + discord.EndpointUserAvatar(user.ID.String(), user.Avatar)
}

func GetGuildMemberDisplayName(member discord.GuildMember) string {
	if member.Nick != "" {
		return member.Nick
	}

	return GetUserDisplayName(member.User)
}

func GetUserDisplayName(user *discord.User) string {
	if user.GlobalName != "" {
		return user.GlobalName
	}

	if user.Discriminator != "" && user.Discriminator != "0" {
		return user.Username + "#" + user.Discriminator
	}

	return user.Username
}

// StubUser represents a user on discord.
type StubUser struct {
	ID   discord.Snowflake `json:"id"`
	Name string            `json:"name"`

	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	GlobalName    string `json:"global_name"`

	Mention string `json:"mention"`

	CreatedAt StubTime `json:"created_at"`
	JoinedAt  StubTime `json:"joined_at"`

	Avatar  string `json:"avatar"`
	Bot     bool   `json:"bot"`
	Pending bool   `json:"pending"`
}

func (s StubUser) String() string {
	if s.GlobalName != "" {
		return s.GlobalName
	}

	if s.Discriminator != "" && s.Discriminator != "0" {
		return s.Username + "#" + s.Discriminator
	}

	return s.Username
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

func (s StubGuild) String() string {
	return s.Name
}

type StubTime time.Time

func (s StubTime) String() string {
	return s.Relative()
}

func (s StubTime) Relative() string {
	return "<t:" + Itoa(time.Time(s).Unix()) + ":R>"
}
