package database

import (
	"context"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
)

// Returns all valid user memberships based on a guild ID.
func (q *Queries) GetValidUserMembershipsByGuildID(ctx context.Context, guildID discord.Snowflake, time time.Time) ([]*GetUserMembershipsByGuildIDRow, error) {
	userMemberships, err := q.GetUserMembershipsByGuildID(ctx, int64(guildID))
	if err != nil {
		return []*GetUserMembershipsByGuildIDRow{}, err
	}

	filteredUserMemberships := make([]*GetUserMembershipsByGuildIDRow, 0, len(userMemberships))

	for _, membership := range userMemberships {
		if isUserMembershipValid(membership, time) {
			filteredUserMemberships = append(filteredUserMemberships, membership)
		}
	}

	return filteredUserMemberships, nil
}

// Returns if a user membership (and transaction) is valid.
func isUserMembershipValid(userMembership *GetUserMembershipsByGuildIDRow, time time.Time) bool {
	// Validate membership has already started and has not ended.
	if userMembership.StartedAt.After(time) || userMembership.ExpiresAt.Before(time) {
		return false
	}

	// Validate the membership status is active.
	if MembershipStatus(userMembership.Status) != MembershipStatusActive {
		return false
	}

	if !userMembership.TransactionUuid.IsNil() {
		// Validate transaction is in database if referenced.
		if !userMembership.TransactionUuid_2.Valid {
			return false
		}

		// Validate transaction is completed.
		if !userMembership.TransactionStatus.Valid {
			return false
		}

		if TransactionStatus(userMembership.TransactionStatus.Int32) != TransactionStatusCompleted {
			return false
		}
	}

	return true
}
