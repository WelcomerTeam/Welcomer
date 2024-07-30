package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image/gif"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/joho/godotenv/autoload"
	"gopkg.in/yaml.v3"
)

type ShareServiceStructure struct {
	MaxValue int64 `json:"max_value"`
	Rows     []struct {
		ID    int64  `json:"id"`
		Value string `json:"value"`
	} `json:"rows"`
}

type ShareServiceStructureUUID struct {
	MaxValue uuid.UUID `json:"max_value"`
	Rows     []struct {
		ID    uuid.UUID `json:"id"`
		Value string    `json:"value"`
	} `json:"rows"`
}

type StringNumber string

var fixes = map[string]string{
	"{user}":                     "{{User.Name}}",
	"{user.mention}":             "{{User.Mention}}",
	"{user.name}":                "{{User.Name}}",
	"{user.discriminator}":       "{{User.Discriminator}}",
	"{user.id}":                  "{{User.ID}}",
	"{user.avatar}":              "{{User.Avatar}}",
	"{user.created.timestamp}":   "{{User.CreatedAt}}",
	"{user.created.since}":       "{{User.CreatedAt}}",
	"{user.joined.timestamp}":    "{{User.JoinedAt}}",
	"{user.joined.since}":        "{{User.JoinedAt}}",
	"{members}":                  "{{Ordinal(Guild.Members)}}",
	"{server}":                   "{{Guild.Name}}",
	"{server.name}":              "{{Guild.Name}}",
	"{server.id}":                "{{Guild.ID}}",
	"{server.members}":           "{{Ordinal(Guild.Members)}}",
	"{server.member.count}":      "{{Guild.Members}}",
	"{server.member.prefix}":     "{{Ordinal(Guild.Members)}}",
	"{server.icon}":              "{{Guild.Icon}}",
	"{server.created.timestamp}": "",
	"{server.created.since}":     "",
	"{server.splash}":            "{{Guild.Splash}}",
	"{server.shard_id}":          "",
	"{invite.code}":              "{{Invite.Code}}",
	"{invite.inviter}":           "{{Invite.Inviter.Name}}",
	"{invite.inviter.id}":        "{{Invite.Inviter.ID}}",
	"{invite.uses}":              "{{Invite.Uses}}",
	"{invite.temporary}":         "{{Invite.Temporary}}",
	"{invite.created.timestamp}": "{{Invite.CreatedAt}}",
	"{invite.created.since}":     "{{Invite.CreatedAt}}",
	"{invite.max}":               "{{Invite.Max}}",
	"{link}":                     "{{Borderwall.Link}}",
	"{guild}":                    "{{Guild.Name}}",
	"{mention}":                  "{{User.Mention}}",
	"{id}":                       "{{User.ID}}",
}

func fixStrings(i string) string {
	for k, v := range fixes {
		i = strings.ReplaceAll(i, k, v)
	}
	return i
}

func fixFormats(i pgtype.JSONB) pgtype.JSONB {
	s := string(i.Bytes)

	for k, v := range fixes {
		s = strings.ReplaceAll(s, k, v)
	}

	i.Bytes = []byte(s)
	return i
}

func (s *StringNumber) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	switch v.(type) {
	case string:
		*s = StringNumber(v.(string))
	case float64:
		*s = StringNumber(strconv.FormatFloat(v.(float64), 'f', -1, 64))
	}

	return nil
}

type PossibleStringList []StringNumber

func (s *PossibleStringList) NewLines() string {
	a := make([]string, len(*s))
	for i, v := range *s {
		a[i] = string(v)
	}
	return strings.Join(a, "\n")
}

func (s *PossibleStringList) UnmarshalJSON(data []byte) error {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	switch v.(type) {
	case string:
		*s = PossibleStringList{StringNumber(v.(string))}
	case []interface{}:
		arr := v.([]interface{})
		l := make(PossibleStringList, len(arr))
		for i, item := range arr {
			switch item.(type) {
			case string:
				l[i] = StringNumber(item.(string))
			}
		}
		*s = l
	}

	return nil
}

func (s *StringNumber) AsInt64() int64 {
	i, _ := strconv.ParseInt(string(*s), 10, 64)
	return i
}

func m(a any) string {
	r, _ := json.Marshal(a)
	return string(r)
}

func PossibleStringListAsInt64(l PossibleStringList) []int64 {
	v := make([]int64, len(l))
	for i, s := range l {
		v[i] = s.AsInt64()
	}
	return v
}

func StringNumberListAsInt64(l []StringNumber) []int64 {
	v := make([]int64, len(l))
	for i, s := range l {
		v[i] = s.AsInt64()
	}
	return v
}

// Embed represents a message embed.
type Embed struct {
	Content     string                  `json:"content,omitempty" yaml:"content,omitempty"`
	Video       *discord.EmbedVideo     `json:"video,omitempty" yaml:"video,omitempty"`
	Timestamp   *time.Time              `json:"timestamp,omitempty" yaml:"timestamp,omitempty"`
	Footer      interface{}             `json:"footer,omitempty" yaml:"footer,omitempty"`
	Image       *discord.EmbedImage     `json:"image,omitempty" yaml:"image,omitempty"`
	Thumbnail   *discord.EmbedThumbnail `json:"thumbnail,omitempty" yaml:"thumbnail,omitempty"`
	Provider    *discord.EmbedProvider  `json:"provider,omitempty" yaml:"provider,omitempty"`
	Author      *discord.EmbedAuthor    `json:"author,omitempty" yaml:"author,omitempty"`
	Type        discord.EmbedType       `json:"type,omitempty" yaml:"type,omitempty"`
	Description string                  `json:"description,omitempty" yaml:"description,omitempty"`
	URL         string                  `json:"url,omitempty" yaml:"url,omitempty"`
	Title       string                  `json:"title,omitempty" yaml:"title,omitempty"`
	Fields      []*discord.EmbedField   `json:"fields,omitempty" yaml:"fields,omitempty"`
	Color       int32                   `json:"color,omitempty" yaml:"color,omitempty"`
}

type NewEmbed struct {
	Content string          `json:"content,omitempty"`
	Embeds  []discord.Embed `json:"embeds,omitempty"`
}

func funnyWelcomerEmbedToEmbedJSON(v []string) pgtype.JSONB {
	var e Embed
	var ne NewEmbed
	var b []byte

	if len(v) == 0 {
		return pgtype.JSONB{
			Bytes:  []byte("{}"),
			Status: pgtype.Present,
		}
	}

	if len(v) == 1 {
		return StringToContentJSONB(v[0])
	}

	format := v[0]
	content := v[1]

	switch format {
	case "y":
		err := yaml.Unmarshal([]byte(content), &e)
		if err != nil {
			e.Content = content
		}

		var footer discord.EmbedFooter

		switch e.Footer.(type) {
		case string:
			footer.Text = e.Footer.(string)
		case map[interface{}]interface{}:
			footerMap := e.Footer.(map[interface{}]interface{})
			footer.Text = footerMap["text"].(string)
			footer.IconURL = footerMap["icon_url"].(string)
			footer.ProxyIconURL = footerMap["proxy_icon_url"].(string)
		}

		ne.Content = e.Content
		ne.Embeds = []discord.Embed{
			{
				Video:       e.Video,
				Timestamp:   e.Timestamp,
				Footer:      &footer,
				Image:       e.Image,
				Thumbnail:   e.Thumbnail,
				Provider:    e.Provider,
				Author:      e.Author,
				Type:        e.Type,
				Description: e.Description,
				URL:         e.URL,
				Title:       e.Title,
				Fields:      e.Fields,
				Color:       e.Color,
			},
		}

		b, err = json.Marshal(e)
		if err != nil {
			b = nil
		}

	case "j":

		fixes := map[string]string{
			"\"url\":false": "\"url\":null",
		}

		for k, v := range fixes {
			content = strings.ReplaceAll(content, k, v)
		}

		err := json.Unmarshal([]byte(content), &e)
		if err != nil {
			content = strings.ReplaceAll(content, "\n", "\\n")
			err = json.Unmarshal([]byte(content), &e)
		}

		if err != nil {
			if strings.HasPrefix(content, "{") && errors.Is(err, &json.UnmarshalTypeError{}) {
				println("Cannot unmarshal json", content, err.Error())
			} else {
				e.Content = content
			}
		} else {

			var footer discord.EmbedFooter

			switch e.Footer.(type) {
			case string:
				footer.Text = e.Footer.(string)
			case map[interface{}]interface{}:
				footerMap := e.Footer.(map[interface{}]interface{})
				footer.Text = footerMap["text"].(string)
				footer.IconURL = footerMap["icon_url"].(string)
				footer.ProxyIconURL = footerMap["proxy_icon_url"].(string)
			}

			ne.Content = e.Content
			ne.Embeds = []discord.Embed{
				{
					Video:       e.Video,
					Timestamp:   e.Timestamp,
					Footer:      &footer,
					Image:       e.Image,
					Thumbnail:   e.Thumbnail,
					Provider:    e.Provider,
					Author:      e.Author,
					Type:        e.Type,
					Description: e.Description,
					URL:         e.URL,
					Title:       e.Title,
					Fields:      e.Fields,
					Color:       e.Color,
				},
			}

			b, err = json.Marshal(e)
			if err != nil {
				b = nil
			}
		}
	default:
		panic("Unknown format" + format)
	}

	r := pgtype.JSONB{}
	if b == nil {
		r.Bytes = []byte("{}")
		r.Status = pgtype.Present
	} else {
		r.Bytes = b
		r.Status = pgtype.Present
	}

	return r
}

type Content struct {
	Content string `json:"content"`
}

type Embeds struct {
	Embeds []struct {
		Description string `json:"description"`
	} `json:"embeds"`
}

func StringToContentJSONB(m string) pgtype.JSONB {
	r := pgtype.JSONB{}
	c := Content{Content: m}
	b, _ := json.Marshal(c)
	r.Bytes = b
	r.Status = pgtype.Present
	return r
}

func StringToEmbedJSONB(m string) pgtype.JSONB {
	r := pgtype.JSONB{}
	c := Embeds{
		Embeds: []struct {
			Description string `json:"description"`
		}{
			{Description: m},
		},
	}
	b, _ := json.Marshal(c)
	r.Bytes = b
	r.Status = pgtype.Present
	return r
}

type GuildInfo struct {
	ID string `json:"id"`

	Dashboard struct {
		DataVersion int `json:"dv"` // guildinfo.details.data_version

		General struct {
			GuildCreation int            `json:"gc"` // guildinfo.details.general.guild_creation
			GuildAddition int            `json:"ga"` // guildinfo.details.general.guild_addition
			Icon          []StringNumber `json:"i"`  // guildinfo.details.general.icon
			Name          StringNumber   `json:"n"`  // guildinfo.details.general.name
			Owner         StringNumber   `json:"o"`  // guildinfo.details.general.owner
			ForceDms      bool           `json:"fd"` // guildinfo.details.general.force_dms
		} `json:"g"`

		Invites []struct {
			Max           int          `json:"max"`
			Code          string       `json:"code"`
			Temp          bool         `json:"temp"`
			Uses          int64        `json:"uses"`
			Channel       StringNumber `json:"channel"`
			Inviter       StringNumber `json:"inviter"`
			Duration      StringNumber `json:"duration"`
			CreatedAt     int64        `json:"created_at,omitempty"`
			ChannelString string       `json:"channel_str"`
			InviterString string       `json:"inviter_str"`
		} `json:"i"`

		Bot struct {
			Cluster     StringNumber `json:"c"`  // guildinfo.details.bot.cluster
			HasDonator  bool         `json:"hd"` // guildinfo.details.bot.has_donator
			HasCustomBg bool         `json:"hb"` // guildinfo.details.bot.has_custombg
			Shard       StringNumber `json:"sh"` // guildinfo.details.bot.shard
			Locale      StringNumber `json:"l"`  // guildinfo.details.bot.locale
			Splash      string       `json:"s"`  // guildinfo.details.bot.splash
			ShowStaff   bool         `json:"ss"` // guildinfo.details.bot.show_staff
			AllowInvite bool         `json:"ai"` // guildinfo.details.bot.allow_invite
			Prefix      StringNumber `json:"p"`  // guildinfo.details.bot.prefix
			Description StringNumber `json:"d"`  // guildinfo.details.bot.description
			EmbedColour StringNumber `json:"ec"` // guildinfo.details.bot.embed_colour
		} `json:"b"`

		Donations []StringNumber `json:"de"` // guildinfo.details.donations
	} `json:"d"`

	Analytics struct {
		Enabled bool `json:"e"` // guildinfo.analytics.enabled
	} `json:"a"`

	Rules struct {
		Rules   []string `json:"r"` // guildinfo.rules.rules
		Enabled bool     `json:"e"` // guildinfo.rules.enabled
	} `json:"r"`

	Channels struct {
		Black struct {
			Enabled  bool           `json:"e"` // guildinfo.channels.black.enabled
			Channels []StringNumber `json:"c"` // guildinfo.channels.black.channels
		} `json:"b"`
		White struct {
			Enabled  bool           `json:"e"` // guildinfo.channels.white.enabled
			Channels []StringNumber `json:"c"` // guildinfo.channels.white.channels
		} `json:"w"`
		Bypass bool `json:"by"` // guildinfo.channels.bypass
	} `json:"ch"`

	BorderwallWall struct {
		Channel         StringNumber       `json:"c"`  // guildinfo.borderwall_wall.channel
		DirectMessage   bool               `json:"d"`  // guildinfo.borderwall_wall.direct_message
		Enabled         bool               `json:"e"`  // guildinfo.borderwall_wall.enabled
		MessageVerify   string             `json:"m"`  // guildinfo.borderwall_wall.message_verify
		Roles           PossibleStringList `json:"r"`  // guildinfo.borderwall_wall.roles
		MessageVerified string             `json:"mv"` // guildinfo.borderwall_wall.message_verified
	} `json:"bw"`

	ServerLock struct {
		Whitelist struct {
			Enabled bool           `json:"e"` // guildinfo.server_lock.whitelist.enabled
			Users   []StringNumber `json:"u"` // guildinfo.server_lock.whitelist.users
		} `json:"wl"`
		Pass struct {
			Enabled bool   `json:"e"` // guildinfo.server_lock.pass.enabled
			Hash    string `json:"h"` // guildinfo.server_lock.pass.hash
		} `json:"p"`
	} `json:"sl"`

	TempChannel struct {
		Enabled      bool         `json:"e"`  // guildinfo.tempchannel.enabled
		AutoPurge    bool         `json:"ap"` // guildinfo.tempchannel.auto_purge
		Category     StringNumber `json:"c"`  // guildinfo.tempchannel.category
		DefaultLimit int          `json:"dl"` // guildinfo.tempchannel.default_limit
		Lobby        StringNumber `json:"l"`  // guildinfo.tempchannel.lobby
	} `json:"tc"`

	AutoRole struct {
		Enabled bool           `json:"e"` // guildinfo.autorole.enabled
		Roles   []StringNumber `json:"r"` // guildinfo.autorole.roles
	} `json:"ar"`

	Leaver struct {
		Enabled  bool         `json:"e"`  // guildinfo.leaver.enabled
		Channel  StringNumber `json:"c"`  // guildinfo.leaver.channel
		Embed    bool         `json:"em"` // guildinfo.leaver.embed
		UseEmbed bool         `json:"ue"` // guildinfo.leaver.useembed
		Text     string       `json:"t"`  // guildinfo.leaver.text
	} `json:"l"`

	FreeRole struct {
		Enabled bool           `json:"e"` // guildinfo.freerole.enabled
		Roles   []StringNumber `json:"r"` // guildinfo.freerole.roles
	} `json:"fr"`

	TimeRoles struct {
		Enabled bool             `json:"e"` // guildinfo.timeroles.enabled
		Roles   [][]StringNumber `json:"r"` // guildinfo.timeroles.roles
	} `json:"tr"`

	NamePurge struct {
		Enabled    bool     `json:"e"` // guildinfo.name_purge.enabled
		IgnoreBots bool     `json:"i"` // guildinfo.name_purge.ignore_bots
		Filter     []string `json:"f"` // guildinfo.name_purge.filter
	} `json:"np"`

	Welcomer struct {
		Channel   StringNumber `json:"c"`  // guildinfo.welcomer.channel
		UseEmbeds bool         `json:"e"`  // guildinfo.welcomer.embed
		Badges    bool         `json:"b"`  // guildinfo.welcomer.badges
		Invites   bool         `json:"iv"` // guildinfo.welcomer.invited

		Images struct {
			Enabled    bool         `json:"e"`  // guildinfo.welcomer.images.enabled
			Background StringNumber `json:"bg"` // guildinfo.welcomer.images.background
			Border     bool         `json:"b"`  // guildinfo.welcomer.images.border
			Colours    struct {
				TextBorder    StringNumber `json:"bo"` // guildinfo.welcomer.images.colours.text_border
				TextBack      StringNumber `json:"b"`  // guildinfo.welcomer.images.colours.text_back
				ProfileBorder StringNumber `json:"pb"` // guildinfo.welcomer.images.colours.profile_border
				ImageBorder   StringNumber `json:"ib"` // guildinfo.welcomer.images.colours.image_border
			} `json:"c"`
			Align   StringNumber       `json:"a"` // guildinfo.welcomer.images.align
			Theme   StringNumber       `json:"t"` // guildinfo.welcomer.images.theme
			Message PossibleStringList `json:"m"` // guildinfo.welcomer.images.message
		} `json:"i"`

		Text struct {
			Enabled bool   `json:"e"` // guildinfo.welcomer.text.enabled
			Message string `json:"m"` // guildinfo.welcomer.text.message
		} `json:"t"`

		Embed          []string `json:"em"` // guildinfo.welcomer.embed
		UseCustomEmbed bool     `json:"ue"` // guildinfo.welcomer.useembed

		DirectMessage struct {
			Enabled  bool     `json:"e"`  // guildinfo.welcomer.dm.enabled
			Message  string   `json:"m"`  // guildinfo.welcomer.dm.message
			Embed    []string `json:"em"` // guildinfo.welcomer.dm.embed
			UseEmbed bool     `json:"ue"` // guildinfo.welcomer.dm.useembed
		} `json:"dm"`
	} `json:"w"`
}

type UserInfo struct {
	Badges  [][]string `json:"b"`
	General struct {
		Bot struct {
			Locale     string `json:"l"`
			Background string `json:"bg"`
			HideMutual bool   `json:"hm"`
			PreferDMs  bool   `json:"pd"`
		} `json:"b"`
	} `json:"g"`
	Memberships struct {
		WelcomerPro1 struct {
			Has      bool    `json:"h"`
			Pledging bool    `json:"p"`
			Until    float64 `json:"u"`
		} `json:"1"`
		WelcomerPro3 struct {
			Has      bool    `json:"h"`
			Pledging bool    `json:"p"`
			Until    float64 `json:"u"`
		} `json:"3"`
		WelcomerPro5 struct {
			Has      bool    `json:"h"`
			Pledging bool    `json:"p"`
			Until    float64 `json:"u"`
		} `json:"5"`
		IsPartner     bool `json:"p"`
		Subscriptions []struct {
			GuildID StringNumber `json:"id"`
			Type    string       `json:"type"`
		} `json:"s"`
		UpdatedAt            int  `json:"u,omitempty"`
		HasCustomBackgrounds bool `json:"hs"`
	} `json:"m"`
	Rep struct {
		LastRep float64 `json:"l"`
		Rep     int     `json:"r"`
	} `json:"r"`
	ID string   `json:"id"`
	IP []string `json:"ip"`
}

type BorderwallInfo struct {
	Activated bool         `json:"a"`  // borderwall.activated
	GuildID   StringNumber `json:"gi"` // borderwall.guild_id
	UserID    StringNumber `json:"ui"` // borderwall.user_id
	IPAddress string       `json:"ip"` // borderwall.ip_address
}

func migrateGuildData(id int64, structure GuildInfo) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in migrateGuildData", r)
		}
	}()

	log := func(args ...interface{}) {
		v := make([]interface{}, 0)
		v = append(v, fmt.Sprintf("[Guild %d]", id))
		v = append(v, args...)
		fmt.Println(v...)
	}

	// Settings
	_, err := q.CreateOrUpdateGuild(ctx, database.CreateOrUpdateGuildParams{
		GuildID:          id,
		EmbedColour:      int32(structure.Dashboard.Bot.EmbedColour.AsInt64()),
		SiteSplashUrl:    structure.Dashboard.Bot.Splash,
		SiteStaffVisible: structure.Dashboard.Bot.ShowStaff,
		SiteGuildVisible: false,
		SiteAllowInvites: structure.Dashboard.Bot.AllowInvite,
	})
	if err != nil {
		log("Cannot create or update guild", id, err.Error())
	}

	// Migrate invites
	invites, err := q.GetGuildInvites(ctx, id)
	if err != nil {
		log("Cannot get guild invites", id, err.Error())
	} else {
		for _, invite := range invites {
			_, err := q.DeleteGuildInvites(ctx, database.DeleteGuildInvitesParams{
				InviteCode: invite.InviteCode,
				GuildID:    id,
			})
			if err != nil {
				log("Cannot delete guild invites", id, invite.InviteCode, err.Error())
			}
		}
	}

	for _, invite := range structure.Dashboard.Invites {
		_, err := q.CreateOrUpdateGuildInvites(ctx, database.CreateOrUpdateGuildInvitesParams{
			InviteCode: invite.Code,
			GuildID:    id,
			CreatedBy:  invite.Inviter.AsInt64(),
			CreatedAt:  time.Unix(invite.CreatedAt, 0),
			Uses:       invite.Uses,
		})
		if err != nil {
			log("Cannot create or update guild invites", id, invite.Code, err.Error())
		}
	}

	// Autoroles
	_, err = q.CreateOrUpdateAutoRolesGuildSettings(ctx, database.CreateOrUpdateAutoRolesGuildSettingsParams{
		GuildID:       id,
		ToggleEnabled: structure.AutoRole.Enabled,
		Roles:         StringNumberListAsInt64(structure.AutoRole.Roles),
	})
	if err != nil {
		log("Cannot create or update auto roles guild settings", id, err.Error())
	}

	// Borderwall
	_, err = q.CreateOrUpdateBorderwallGuildSettings(ctx, database.CreateOrUpdateBorderwallGuildSettingsParams{
		GuildID:         id,
		ToggleEnabled:   structure.BorderwallWall.Enabled,
		ToggleSendDm:    structure.BorderwallWall.DirectMessage,
		Channel:         structure.BorderwallWall.Channel.AsInt64(),
		MessageVerify:   fixFormats(StringToContentJSONB(structure.BorderwallWall.MessageVerify)),
		MessageVerified: fixFormats(StringToContentJSONB(structure.BorderwallWall.MessageVerified)),
		RolesOnJoin:     []int64{},
		RolesOnVerify:   PossibleStringListAsInt64(structure.BorderwallWall.Roles),
	})
	if err != nil {
		log("Cannot create or update borderwall guild settings", id, err.Error())
	}

	// Freeroles
	_, err = q.CreateOrUpdateFreeRolesGuildSettings(ctx, database.CreateOrUpdateFreeRolesGuildSettingsParams{
		GuildID:       id,
		ToggleEnabled: structure.FreeRole.Enabled,
		Roles:         StringNumberListAsInt64(structure.FreeRole.Roles),
	})
	if err != nil {
		log("Cannot create or update free roles guild settings", id, err.Error())
	}

	// Leavers
	var leaverMessageFormat pgtype.JSONB

	if structure.Leaver.Embed {
		leaverMessageFormat = StringToEmbedJSONB(structure.Leaver.Text)
	} else {
		leaverMessageFormat = StringToContentJSONB(structure.Leaver.Text)
	}

	_, err = q.CreateOrUpdateLeaverGuildSettings(ctx, database.CreateOrUpdateLeaverGuildSettingsParams{
		GuildID:       id,
		ToggleEnabled: structure.Leaver.Enabled,
		Channel:       structure.Leaver.Channel.AsInt64(),
		MessageFormat: fixFormats(leaverMessageFormat),
	})
	if err != nil {
		log("Cannot create or update leaver guild settings", id, err.Error())
	}

	// Rules
	_, err = q.CreateOrUpdateRulesGuildSettings(ctx, database.CreateOrUpdateRulesGuildSettingsParams{
		GuildID:          id,
		ToggleEnabled:    structure.Rules.Enabled,
		ToggleDmsEnabled: false,
		Rules:            structure.Rules.Rules,
	})
	if err != nil {
		log("Cannot create or update rules guild settings", id, err.Error())
	}

	// TempChannels
	_, err = q.CreateOrUpdateTempChannelsGuildSettings(ctx, database.CreateOrUpdateTempChannelsGuildSettingsParams{
		GuildID:          id,
		ToggleEnabled:    structure.TempChannel.Enabled,
		ToggleAutopurge:  structure.TempChannel.AutoPurge,
		ChannelLobby:     structure.TempChannel.Lobby.AsInt64(),
		ChannelCategory:  structure.TempChannel.Category.AsInt64(),
		DefaultUserCount: int32(structure.TempChannel.DefaultLimit),
	})
	if err != nil {
		log("Cannot create or update temp channels guild settings", id, err.Error())
	}

	// TimeRoles
	timeRoles := []welcomer.GuildSettingsTimeRolesRole{}
	for _, role := range structure.TimeRoles.Roles {
		timeRoles = append(timeRoles, welcomer.GuildSettingsTimeRolesRole{
			Role:    discord.Snowflake(role[0].AsInt64()),
			Seconds: int(role[1].AsInt64()),
		})
	}

	timeRolesJSON := pgtype.JSONB{}
	timeRolesJSON.Bytes, _ = json.Marshal(timeRoles)
	timeRolesJSON.Status = pgtype.Present

	_, err = q.CreateOrUpdateTimeRolesGuildSettings(ctx, database.CreateOrUpdateTimeRolesGuildSettingsParams{
		GuildID:       id,
		ToggleEnabled: structure.TimeRoles.Enabled,
		Timeroles:     timeRolesJSON,
	})
	if err != nil {
		log("Cannot create or update time roles guild settings", id, err.Error())
	}

	// Welcomer
	var welcomerMessageFormat pgtype.JSONB
	if structure.Welcomer.UseCustomEmbed {
		welcomerMessageFormat = funnyWelcomerEmbedToEmbedJSON(structure.Welcomer.Embed)
	} else {
		if structure.Welcomer.UseEmbeds {
			welcomerMessageFormat = StringToEmbedJSONB(structure.Welcomer.Text.Message)
		} else {
			welcomerMessageFormat = StringToContentJSONB(structure.Welcomer.Text.Message)
		}
	}

	_, err = q.CreateOrUpdateWelcomerTextGuildSettings(ctx, database.CreateOrUpdateWelcomerTextGuildSettingsParams{
		GuildID:       id,
		ToggleEnabled: structure.Welcomer.Text.Enabled,
		Channel:       structure.Welcomer.Channel.AsInt64(),
		MessageFormat: fixFormats(welcomerMessageFormat),
	})
	if err != nil {
		log("Cannot create or update welcomer text guild settings", id, err.Error())
	}

	fixColour := func(i string) string {
		if i == "" || i == "0" {
			return "#000000"
		}

		o := strings.ReplaceAll(strings.ReplaceAll(i, "RGB|", "#"), "RGBA|", "#")
		if !strings.HasPrefix(o, "#") {
			o = "#" + o
		}

		return o
	}

	if strings.HasPrefix(string(structure.Welcomer.Images.Background), "custom_") {
		req, err := http.NewRequest("GET", os.Getenv("SHARE_SERVICE_URL")+"/fetch_background?name="+string(structure.Welcomer.Images.Background), nil)
		if err != nil {
			panic(err)
		}

		req.Header.Add("Authorization", "Bearer "+os.Getenv("SHARE_SERVICE_TOKEN"))
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}

			var welcomerImageUuid uuid.UUID

			var gen = uuid.NewGen()
			welcomerImageUuid, _ = gen.NewV7()

			var imageType string

			_, err = gif.Decode(bytes.NewBuffer(respBody))
			if err != nil {
				imageType = utils.ImageFileTypeImagePng.String()
			} else {
				imageType = utils.ImageFileTypeImageGif.String()
			}

			im, _ := q.GetWelcomerImagesByGuildId(ctx, id)
			for _, i := range im {
				_, err = q.DeleteWelcomerImage(ctx, i.ImageUuid)
				if err != nil {
					log("Cannot delete welcomer images", id, err.Error())
				}
			}

			_, err = q.CreateWelcomerImages(ctx, database.CreateWelcomerImagesParams{
				ImageUuid: welcomerImageUuid,
				GuildID:   id,
				CreatedAt: time.Now(),
				ImageType: imageType,
				Data:      respBody,
			})
			if err != nil {
				log("Cannot create welcomer images", id, err.Error())
				structure.Welcomer.Images.Background = "default"
			} else {
				structure.Welcomer.Images.Background = "custom:" + StringNumber(welcomerImageUuid.String())
			}
		} else {
			structure.Welcomer.Images.Background = "default"
		}
	}

	imageAlignmentMapping := map[int64]int32{
		0: int32(utils.ImageAlignmentLeft),   // Left
		1: int32(utils.ImageAlignmentCenter), // Center
		2: int32(utils.ImageAlignmentRight),  // Right
	}

	imageThemeMapping := map[int64]int32{
		0: int32(utils.ImageThemeDefault),  // Legacy
		1: int32(utils.ImageThemeVertical), // Vertical
		2: int32(utils.ImageThemeDefault),  // Badge
		3: int32(utils.ImageThemeDefault),  // Shadowed
		4: int32(utils.ImageThemeDefault),  // Widget
	}

	imageProfileBorderMapping := map[int64]int32{
		0: int32(utils.ImageProfileBorderTypeCircular), // Circular
		1: int32(utils.ImageProfileBorderTypeSquared),  // Square
		2: int32(utils.ImageProfileBorderTypeSquared),  // Square
		3: int32(utils.ImageProfileBorderTypeCircular), // Circular
		4: 4,                                           // None
	}

	_, err = q.CreateOrUpdateWelcomerImagesGuildSettings(ctx, database.CreateOrUpdateWelcomerImagesGuildSettingsParams{
		GuildID:                id,
		ToggleEnabled:          structure.Welcomer.Images.Enabled,
		ToggleImageBorder:      structure.Welcomer.Images.Border,
		BackgroundName:         string(structure.Welcomer.Images.Background),
		ColourText:             fixColour(string(structure.Welcomer.Images.Colours.TextBack)),
		ColourTextBorder:       fixColour(string(structure.Welcomer.Images.Colours.TextBorder)),
		ColourImageBorder:      fixColour(string(structure.Welcomer.Images.Colours.ImageBorder)),
		ColourProfileBorder:    fixColour(string(structure.Welcomer.Images.Colours.ProfileBorder)),
		ImageAlignment:         imageAlignmentMapping[structure.Welcomer.Images.Align.AsInt64()],
		ImageTheme:             imageThemeMapping[structure.Welcomer.Images.Theme.AsInt64()],
		ImageMessage:           fixStrings(structure.Welcomer.Images.Message.NewLines()),
		ImageProfileBorderType: imageProfileBorderMapping[structure.Welcomer.Images.Colours.ProfileBorder.AsInt64()],
	})
	if err != nil {
		log("Cannot create or update welcomer images guild settings", id, err.Error())
	}

	var welcomerDMSMessageFormat pgtype.JSONB
	if structure.Welcomer.DirectMessage.UseEmbed {
		welcomerDMSMessageFormat = funnyWelcomerEmbedToEmbedJSON(structure.Welcomer.DirectMessage.Embed)
	} else {
		welcomerDMSMessageFormat = StringToContentJSONB(structure.Welcomer.DirectMessage.Message)
	}

	_, err = q.CreateOrUpdateWelcomerDMsGuildSettings(ctx, database.CreateOrUpdateWelcomerDMsGuildSettingsParams{
		GuildID:             id,
		ToggleEnabled:       structure.Welcomer.DirectMessage.Enabled,
		ToggleUseTextFormat: false,
		ToggleIncludeImage:  false,
		MessageFormat:       fixFormats(welcomerDMSMessageFormat),
	})
	if err != nil {
		log("Cannot create or update welcomer guild settings", id, err.Error())
	}
}

func migrateUserData(id int64, structure UserInfo) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in migrateUserData", r)
		}
	}()

	nowUnix := time.Now().Unix()
	// Create transactions

	m, err := q.GetUserMembershipsByUserID(ctx, id)
	if err != nil {
		println("Cannot get user memberships", id, err.Error())
	} else {
		for _, i := range m {
			_, err := q.DeleteUserMembership(ctx, i.MembershipUuid)
			if err != nil {
				println("Cannot delete membership", id, err.Error())
			}
		}
	}

	t, err := q.GetUserTransactionsByUserID(ctx, id)
	if err != nil {
		println("Cannot get user transactions", id, err.Error())
	} else {
		for _, i := range t {
			_, err := q.DeleteUserTransaction(ctx, i.TransactionUuid)
			if err != nil {
				println("Cannot delete user transaction", id, err.Error())
			}
		}
	}

	memberships := make([]uuid.UUID, 0)
	transactions := make([]uuid.UUID, 0)

	// Welcomer Pro1
	if structure.Memberships.WelcomerPro1.Has {
		t, err := q.CreateUserTransaction(ctx, database.CreateUserTransactionParams{
			UserID:            id,
			PlatformType:      utils.If(structure.Memberships.WelcomerPro1.Pledging, int32(database.PlatformTypePatreon), int32(database.PlatformTypePaypal)),
			TransactionStatus: int32(database.TransactionStatusCompleted),
		})
		if err != nil {
			println("Cannot create welcomer pro1  user transaction", id, err.Error())
		}

		m, err := q.CreateNewMembership(ctx, database.CreateNewMembershipParams{
			StartedAt:       time.Unix(int64(structure.Memberships.WelcomerPro1.Until), 0).AddDate(0, -1, 0),
			ExpiresAt:       time.Unix(int64(structure.Memberships.WelcomerPro1.Until), 0),
			Status:          utils.If(int64(structure.Memberships.WelcomerPro1.Until) > nowUnix, int32(database.MembershipStatusActive), int32(database.MembershipStatusExpired)),
			MembershipType:  int32(database.MembershipTypeLegacyWelcomerPro1),
			TransactionUuid: t.TransactionUuid,
			UserID:          id,
			GuildID:         0,
		})
		if err != nil {
			println("Cannot create new welcomer pro1 membership", id, err.Error())
		} else {
			memberships = append(memberships, m.MembershipUuid)
			transactions = append(transactions, m.TransactionUuid)
		}
	}

	// Welcomer Pro3
	if structure.Memberships.WelcomerPro3.Has {
		t, err := q.CreateUserTransaction(ctx, database.CreateUserTransactionParams{
			UserID:            id,
			PlatformType:      utils.If(structure.Memberships.WelcomerPro3.Pledging, int32(database.PlatformTypePatreon), int32(database.PlatformTypePaypal)),
			TransactionStatus: int32(database.TransactionStatusCompleted),
		})
		if err != nil {
			println("Cannot create welcomer pro1  user transaction", id, err.Error())
		}

		for i := 0; i < 3; i++ {
			m, err := q.CreateNewMembership(ctx, database.CreateNewMembershipParams{
				StartedAt:       time.Unix(int64(structure.Memberships.WelcomerPro3.Until), 0).AddDate(0, -1, 0),
				ExpiresAt:       time.Unix(int64(structure.Memberships.WelcomerPro3.Until), 0),
				Status:          utils.If(int64(structure.Memberships.WelcomerPro3.Until) > nowUnix, int32(database.MembershipStatusActive), int32(database.MembershipStatusExpired)),
				MembershipType:  int32(database.MembershipTypeLegacyWelcomerPro3),
				TransactionUuid: t.TransactionUuid,
				UserID:          id,
				GuildID:         0,
			})
			if err != nil {
				println("Cannot create new welcomer pro1 membership", id, err.Error())
			} else {
				memberships = append(memberships, m.MembershipUuid)
				transactions = append(transactions, m.TransactionUuid)
			}
		}
	}

	// Welcomer Pro5
	if structure.Memberships.WelcomerPro5.Has {
		t, err := q.CreateUserTransaction(ctx, database.CreateUserTransactionParams{
			UserID:            id,
			PlatformType:      utils.If(structure.Memberships.WelcomerPro5.Pledging, int32(database.PlatformTypePatreon), int32(database.PlatformTypePaypal)),
			TransactionStatus: int32(database.TransactionStatusCompleted),
		})
		if err != nil {
			println("Cannot create welcomer pro5 user transaction", id, err.Error())
		}

		for i := 0; i < 5; i++ {
			m, err := q.CreateNewMembership(ctx, database.CreateNewMembershipParams{
				StartedAt:       time.Unix(int64(structure.Memberships.WelcomerPro5.Until), 0).AddDate(0, -1, 0),
				ExpiresAt:       time.Unix(int64(structure.Memberships.WelcomerPro5.Until), 0),
				Status:          utils.If(int64(structure.Memberships.WelcomerPro5.Until) > nowUnix, int32(database.MembershipStatusActive), int32(database.MembershipStatusExpired)),
				MembershipType:  int32(database.MembershipTypeLegacyWelcomerPro5),
				TransactionUuid: t.TransactionUuid,
				UserID:          id,
				GuildID:         0,
			})
			if err != nil {
				println("Cannot create new welcomer pro5 membership", id, err.Error())
			} else {
				memberships = append(memberships, m.MembershipUuid)
				transactions = append(transactions, m.TransactionUuid)
			}
		}
	}

	// Welcomer Partner
	if structure.Memberships.IsPartner {
		t, err := q.CreateUserTransaction(ctx, database.CreateUserTransactionParams{
			UserID:            id,
			PlatformType:      int32(database.PlatformTypeUnknown),
			TransactionStatus: int32(database.TransactionStatusCompleted),
		})
		if err != nil {
			println("Cannot create welcomer parnter user transaction", id, err.Error())
		}

		for i := 0; i < 5; i++ {
			m, err := q.CreateNewMembership(ctx, database.CreateNewMembershipParams{
				StartedAt:       time.Unix(0, 0),
				ExpiresAt:       time.Unix(2147483647, 0),
				Status:          int32(database.MembershipStatusActive),
				MembershipType:  int32(database.MembershipTypeLegacyWelcomerPro5),
				TransactionUuid: t.TransactionUuid,
				UserID:          id,
				GuildID:         0,
			})
			if err != nil {
				println("Cannot create new welcomer parnter membership", id, err.Error())
			} else {
				memberships = append(memberships, m.MembershipUuid)
				transactions = append(transactions, m.TransactionUuid)
			}
		}
	}

	hasCbg := false
	for _, sub := range structure.Memberships.Subscriptions {
		if sub.Type == "cbg" {
			hasCbg = true
			break
		}
	}

	var cbgTx *database.UserTransactions
	if hasCbg {
		cbgTx, err = q.CreateUserTransaction(ctx, database.CreateUserTransactionParams{
			UserID:            id,
			PlatformType:      int32(database.PlatformTypeUnknown),
			TransactionStatus: int32(database.TransactionStatusCompleted),
		})
		if err != nil {
			println("Cannot create cbg user transaction", id, err.Error())
		}
	}

	for _, sub := range structure.Memberships.Subscriptions {
		switch sub.Type {
		case "don":
			if len(transactions) == 0 {
				println("No memberships for donation", id)
			} else {
				v := transactions[0]
				memberships = memberships[1:]
				transactions = transactions[1:]

				_, err := q.CreateNewMembership(ctx, database.CreateNewMembershipParams{
					StartedAt:       time.Unix(0, 0),
					ExpiresAt:       time.Unix(2147483647, 0),
					Status:          int32(database.MembershipStatusActive),
					MembershipType:  int32(database.MembershipTypeLegacyCustomBackgrounds),
					TransactionUuid: v,
					UserID:          id,
					GuildID:         sub.GuildID.AsInt64(),
				})
				if err != nil {
					println("Cannot create new membership", id, err.Error())
				}
			}
		case "cbg":
			_, err := q.CreateNewMembership(ctx, database.CreateNewMembershipParams{
				StartedAt:       time.Unix(0, 0),
				ExpiresAt:       time.Unix(2147483647, 0),
				Status:          int32(database.MembershipStatusActive),
				MembershipType:  int32(database.MembershipTypeLegacyCustomBackgrounds),
				TransactionUuid: cbgTx.TransactionUuid,
				UserID:          id,
				GuildID:         sub.GuildID.AsInt64(),
			})
			if err != nil {
				println("Cannot create new membership", id, err.Error())
			}
		default:
			panic("Unknown subscription type: " + sub.Type)
		}
	}

	for _, remainingSubs := range memberships {
		r, err := q.GetUserMembership(ctx, remainingSubs)
		if err != nil {
			println("Cannot get user membership", id, remainingSubs.String(), err.Error())
		}

		_, err = q.UpdateUserMembership(ctx, database.UpdateUserMembershipParams{
			MembershipUuid:  remainingSubs,
			StartedAt:       time.Time{},
			ExpiresAt:       time.Time{},
			Status:          int32(database.MembershipStatusIdle),
			TransactionUuid: r.TransactionUuid,
			UserID:          r.UserID,
			GuildID:         r.GuildID,
		})
		if err != nil {
			println("Cannot update user membership", id, err.Error())
		}
	}
}

func migrateBorderwallData(id uuid.UUID, structure BorderwallInfo) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in migrateBorderwallData", r)
		}
	}()

	ip := net.ParseIP(structure.IPAddress)
	var pip pgtype.Inet

	if ip == nil {
		pip.Status = pgtype.Null
	} else {
		pip.IPNet = &net.IPNet{
			IP:   ip,
			Mask: ip.DefaultMask(),
		}
		pip.Status = pgtype.Present
	}

	_, err := q.InsertBorderwallRequest(ctx, database.InsertBorderwallRequestParams{
		RequestUuid: id,
		GuildID:     structure.GuildID.AsInt64(),
		UserID:      structure.UserID.AsInt64(),
		IsVerified:  structure.Activated,
		IpAddress:   pip,
	})
	if err != nil {
		println("Cannot migrate borderwall data", id.String(), err.Error())
	}
}

var q *database.Queries
var ctx context.Context
var client http.Client

func main() {
	migrateGuilds := true
	migrateUsers := true
	migrateBorderwall := true

	guildMinValue := int64(839101942046916619)
	userMinValue := int64(0)

	ctx = context.Background()

	pool, err := pgxpool.Connect(ctx, os.Getenv("POSTGRES_URL"))
	if err != nil {
		panic(fmt.Sprintf(`pgxpool.Connect(): %v`, err.Error()))
	}

	q = database.New(pool)

	wg := sync.WaitGroup{}
	client = http.Client{}

	if migrateGuilds {
		min_value := int(guildMinValue)
		total_rows := int(0)

		for {
			req, err := http.NewRequest("GET", os.Getenv("SHARE_SERVICE_URL")+"/guilds?min_value="+strconv.Itoa(min_value), nil)
			if err != nil {
				panic(err)
			}

			req.Header.Add("Authorization", "Bearer "+os.Getenv("SHARE_SERVICE_TOKEN"))
			resp, err := client.Do(req)
			if err != nil {
				panic(err)
			}

			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}

			resp.Body.Close()

			var structure ShareServiceStructure
			err = json.Unmarshal(respBody, &structure)
			if err != nil {
				panic(err)
			}

			min_value = int(structure.MaxValue)
			total_rows += len(structure.Rows)

			wg.Add(1)
			go func(structure ShareServiceStructure) {
				defer wg.Done()
				for _, row := range structure.Rows {
					var guildInfo GuildInfo
					err := json.Unmarshal([]byte(row.Value), &guildInfo)
					if err != nil {
						println("Cannot migrate guild data", row.ID, err.Error())
					} else {
						migrateGuildData(row.ID, guildInfo)
					}
				}
			}(structure)

			println("GUILDS", min_value, total_rows)

			if len(structure.Rows) < 1000 {
				break
			}
		}
	}

	if migrateUsers {
		min_value := int(userMinValue)
		total_rows := int(0)

		for {
			req, err := http.NewRequest("GET", os.Getenv("SHARE_SERVICE_URL")+"/memberships?min_value="+strconv.Itoa(min_value), nil)
			if err != nil {
				panic(err)
			}

			req.Header.Add("Authorization", "Bearer "+os.Getenv("SHARE_SERVICE_TOKEN"))
			resp, err := client.Do(req)
			if err != nil {
				panic(err)
			}

			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}

			resp.Body.Close()

			var structure ShareServiceStructure
			err = json.Unmarshal(respBody, &structure)
			if err != nil {
				panic(err)
			}

			min_value = int(structure.MaxValue)
			total_rows += len(structure.Rows)

			wg.Add(1)
			go func(structure ShareServiceStructure) {
				defer wg.Done()
				for _, row := range structure.Rows {
					var userInfo UserInfo
					err := json.Unmarshal([]byte(row.Value), &userInfo)
					if err != nil {
						println("Cannot migrate user data", row.ID, err.Error(), string(row.Value))
					} else {
						migrateUserData(row.ID, userInfo)
					}
				}
			}(structure)

			println("USERS", min_value, total_rows)

			if len(structure.Rows) < 1000 {
				break
			}
		}
	}

	if migrateBorderwall {
		min_value_uuid := uuid.Nil.String()
		total_rows := int(0)

		for {
			req, err := http.NewRequest("GET", os.Getenv("SHARE_SERVICE_URL")+"/borderwall?min_value="+min_value_uuid, nil)
			if err != nil {
				panic(err)
			}

			req.Header.Add("Authorization", "Bearer "+os.Getenv("SHARE_SERVICE_TOKEN"))
			resp, err := client.Do(req)
			if err != nil {
				panic(err)
			}

			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}

			resp.Body.Close()

			var structure ShareServiceStructureUUID
			err = json.Unmarshal(respBody, &structure)
			if err != nil {
				panic(err)
			}

			min_value_uuid = structure.MaxValue.String()
			total_rows += len(structure.Rows)

			wg.Add(1)
			go func(structure ShareServiceStructureUUID) {
				defer wg.Done()
				for _, row := range structure.Rows {
					var borderwallInfo BorderwallInfo
					err := json.Unmarshal([]byte(row.Value), &borderwallInfo)
					if err != nil {
						println("Cannot migrate borderwall data", row.ID.String(), err.Error())
					} else {
						migrateBorderwallData(row.ID, borderwallInfo)
					}
				}
			}(structure)

			println("BORDERWALL", min_value_uuid, total_rows)

			if len(structure.Rows) < 1000 {
				break
			}
		}
	}

	println("Waiting for all goroutines to finish")
	wg.Wait()
	println("All goroutines finished")
}
