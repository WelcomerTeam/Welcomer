package welcomer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	utils "github.com/WelcomerTeam/Welcomer/welcomer-utils"
)

const (
	PatreonBase = "https://www.patreon.com/api/oauth2/v2/"
	CampaignID  = "1150593"
)

var null = []byte("null")

type PatreonTier int64

const (
	PatreonTierFree PatreonTier = 10503463

	PatreonTierUnpublishedWelcomerDonator PatreonTier = 3975266
	PatreonTierUnpublishedWelcomerPro1    PatreonTier = 3744919
	PatreonTierUnpublishedWelcomerPro3    PatreonTier = 3744921
	PatreonTierUnpublishedWelcomerPro5    PatreonTier = 3744926

	PatreonTierWelcomerPro PatreonTier = 23606682
)

func (s *PatreonTier) UnmarshalJSON(b []byte) error {
	if !bytes.Equal(b, null) {
		i, err := strconv.ParseInt(string(b[1:len(b)-1]), 10, 64)
		if err != nil {
			return fmt.Errorf("failed to unmarshal json: %v", err)
		}

		*s = PatreonTier(i)
	} else {
		*s = 0
	}

	return nil
}

type PatreonMember struct {
	PatreonUserID discord.Snowflake `json:"patreon_user_id"`
	EntitledTiers []PatreonTier     `json:"active_tier"`
	Attributes    Attributes        `json:"attributes"`
}

type GetPatreonMembersResponse struct {
	Data []struct {
		Attributes    Attributes    `json:"attributes"`
		Relationships Relationships `json:"relationships"`
	} `json:"data"`
	Included []struct {
		Attributes Attributes        `json:"attributes"`
		ID         discord.Snowflake `json:"id"`
		Type       string            `json:"type"`
	}
	Links struct {
		Next string `json:"next"`
	} `json:"links"`
}

type Attributes struct {
	CurrentlyEntitledAmountCents int64                   `json:"currently_entitled_amount_cents"`
	Email                        string                  `json:"email"`
	FullName                     string                  `json:"full_name"`
	ThumbUrl                     string                  `json:"thumb_url"`
	IsFollower                   bool                    `json:"is_follower"`
	LastChargeDate               time.Time               `json:"last_charge_date"`
	LastChargeStatus             LastChargeStatus        `json:"last_charge_status"`
	LifetimeSupportCents         int64                   `json:"lifetime_support_cents"`
	PatronStatus                 PatronStatus            `json:"patron_status"`
	SocialConnections            PatronSocialConnections `json:"social_connections"`
}

type PatronSocialConnections struct {
	Discord struct {
		UserID discord.Snowflake `json:"user_id"`
	} `json:"discord"`
}

type LastChargeStatus string

const (
	LastChargeStatusDeclined LastChargeStatus = "Declined"
	LastChargeStatusPaid     LastChargeStatus = "Paid"
	LastChargeStatusDeleted  LastChargeStatus = "Deleted"
	LastChargeStatusPending  LastChargeStatus = "Pending"
	LastChargeStatusRefunded LastChargeStatus = "Refunded"
	LastChargeStatusFraud    LastChargeStatus = "Fraud"
	LastChargeStatusOther    LastChargeStatus = "Other"
)

type PatronStatus string

const (
	PatreonStatusNeverPledged PatronStatus = ""
	PatreonStatusActive       PatronStatus = "active_patron"
	PatreonStatusDeclined     PatronStatus = "declined_patron"
	PatreonStatusFormer       PatronStatus = "former_patron"
)

type Relationships struct {
	CurrentlyEntitledTiers struct {
		Data []struct {
			ID PatreonTier `json:"id"`
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
		PatreonBase+"identity?fields[user]=email,full_name,social_connections,thumb_url",
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

func GetAllPatreonMembers(ctx context.Context, client *http.Client) ([]PatreonMember, error) {
	return getAllPatreonMembers(ctx, client, nil, "")
}

func getAllPatreonMembers(ctx context.Context, client *http.Client, l []PatreonMember, u string) ([]PatreonMember, error) {
	if l == nil {
		l = []PatreonMember{}
	}

	u = utils.If(u != "", u, PatreonBase+"campaigns/"+CampaignID+"/members?fields[member]=patron_status,email,pledge_relationship_start,currently_entitled_amount_cents,last_charge_status,last_charge_date,pledge_relationship_start&fields[user]=social_connections,full_name,thumb_url&include=user,currently_entitled_tiers&page[size]=100")

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		u,
		nil,
	)
	if err != nil {
		return l, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return l, err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return l, fmt.Errorf("received status code %d", resp.StatusCode)
	}

	var response GetPatreonMembersResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return l, err
	}

	for _, member := range response.Data {
		var entitledTiers []PatreonTier

		for _, tier := range member.Relationships.CurrentlyEntitledTiers.Data {
			entitledTiers = append(entitledTiers, tier.ID)
		}

		for _, inc := range response.Included {
			if inc.ID == member.Relationships.User.Data.ID && inc.Type == "user" {
				member.Attributes.FullName = inc.Attributes.FullName
				member.Attributes.SocialConnections.Discord.UserID = inc.Attributes.SocialConnections.Discord.UserID
			}
		}

		l = append(l, PatreonMember{
			PatreonUserID: member.Relationships.User.Data.ID,
			Attributes:    member.Attributes,
			EntitledTiers: entitledTiers,
		})
	}

	if response.Links.Next != "" && response.Links.Next != u {
		l, err = getAllPatreonMembers(ctx, client, l, response.Links.Next)
	}

	return l, err
}
