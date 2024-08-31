package welcomer

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/WelcomerTeam/Discord/discord"
	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
)

const (
	PatreonBase = "https://www.patreon.com/api/oauth2/v2/"
	CampaignID  = "1150593"
)

type PatreonMember struct {
	PatreonUserID discord.Snowflake   `json:"patreon_user_id"`
	EntitledTiers []discord.Snowflake `json:"active_tier"`
}

type GetPatreonMembersResponse struct {
	Data []struct {
		Relationships Relationships `json:"relationships"`
	} `json:"data"`
	Links struct {
		Next string `json:"next"`
	} `json:"links"`
}

type Relationships struct {
	CurrentlyEntitledTiers struct {
		Data []struct {
			ID discord.Snowflake `json:"id"`
		} `json:"data"`
	} `json:"currently_entitled_tiers"`
	User struct {
		Data struct {
			ID discord.Snowflake `json:"id"`
		} `json:"data"`
	} `json:"user"`
}

type PatreonUser struct {
	ID                discord.Snowflake             `json:"id"`
	Email             string                        `json:"email"`
	FullName          string                        `json:"full_name"`
	SocialConnections PatreonUser_SocialConnections `json:"social_connections"`
	ThumbURL          string                        `json:"thumb_url"`
}

type PatreonUserOuter struct {
	Data patreonUser `json:"data"`
}

type patreonUser struct {
	Attributes PatreonUser_Attributes `json:"attributes"`
	ID         discord.Snowflake      `json:"id"`
	Type       string                 `json:"type"`
}

type PatreonUser_Attributes struct {
	Email             string                        `json:"email"`
	FullName          string                        `json:"full_name"`
	SocialConnections PatreonUser_SocialConnections `json:"social_connections"`
	ThumbURL          string                        `json:"thumb_url"`
}

type PatreonUser_SocialConnections struct {
	Discord PatreonUser_Discord `json:"discord"`
}

type PatreonUser_Discord struct {
	UserID discord.Snowflake `json:"user_id"`
}

func IdentifyPatreonMember(ctx context.Context, token string) (PatreonUser, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		PatreonBase+"identity?fields%5Buser%5D=email,full_name,social_connections,thumb_url",
		nil,
	)
	if err != nil {
		return PatreonUser{}, err
	}

	req.Header.Set("Authorization", token)
	req.Header.Set("User-Agent", utils.UserAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return PatreonUser{}, err
	}

	defer resp.Body.Close()

	var user PatreonUserOuter
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return PatreonUser{}, err
	}

	return PatreonUser{
		ID:                user.Data.ID,
		Email:             user.Data.Attributes.Email,
		FullName:          user.Data.Attributes.FullName,
		SocialConnections: user.Data.Attributes.SocialConnections,
		ThumbURL:          user.Data.Attributes.ThumbURL,
	}, nil
}

func GetAllPatreonMembers(ctx context.Context, token string, l []PatreonMember, u string) ([]PatreonMember, error) {
	if l == nil {
		l = []PatreonMember{}
	}

	req, err := http.NewRequest(
		http.MethodGet,
		utils.If(u != "", u, PatreonBase+"campaigns/"+CampaignID+"/members?fields=%5Buser%5D%3Dsocial_connections&include=user%2Ccurrently_entitled_tiers&page%5Bsize%5D=1000"),
		nil,
	)
	if err != nil {
		return l, err
	}

	req.Header.Set("Authorization", token)
	req.Header.Set("User-Agent", utils.UserAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return l, err
	}

	defer resp.Body.Close()

	var response GetPatreonMembersResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return l, err
	}

	for _, member := range response.Data {
		var entitledTiers []discord.Snowflake

		for _, tier := range member.Relationships.CurrentlyEntitledTiers.Data {
			entitledTiers = append(entitledTiers, tier.ID)
		}

		l = append(l, PatreonMember{
			PatreonUserID: member.Relationships.User.Data.ID,
			EntitledTiers: entitledTiers,
		})
	}

	if response.Links.Next != "" {
		l, err = GetAllPatreonMembers(ctx, token, l, response.Links.Next)
	}

	return l, err
}
