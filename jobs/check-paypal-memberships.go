package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/WelcomerTeam/Discord/discord"
	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/WelcomerTeam/Welcomer/welcomer-core/database"
	_ "github.com/joho/godotenv/autoload"
	"github.com/plutov/paypal/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultExpiryWindow = 14 * 24 * time.Hour
	membershipGrace     = 7 * 24 * time.Hour
)

func main() {
	var err error

	loggingLevel := flag.String("level", os.Getenv("LOGGING_LEVEL"), "Logging level")

	postgresURL := flag.String("postgresURL", os.Getenv("POSTGRES_URL"), "Postgres connection URL")
	sandwichGRPCHost := flag.String(
		"sandwichGRPCHost",
		os.Getenv("SANDWICH_GRPC_HOST"),
		"GRPC Address for the Sandwich Daemon service",
	)

	webhookUrl := flag.String(
		"webhookUrl",
		os.Getenv("JOB_WEBHOOK_URL"),
		"Webhook URL for logging",
	)

	sandwichManagerName := flag.String(
		"sandwichManagerName",
		os.Getenv("SANDWICH_MANAGER_NAME"),
		"Sandwich manager identifier name",
	)

	apply := flag.Bool("apply", false, "write membership updates (default is dry-run)")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	ctx = welcomer.AddManagerNameToContext(ctx, "welcomer")

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
			println(string(debug.Stack()))

			err = welcomer.SendWebhookMessage(ctx, *webhookUrl, discord.WebhookMessageParams{
				Content: "<@143090142360371200>",
				Embeds: []discord.Embed{
					{
						Title:       "Check PayPal Memberships Job",
						Description: fmt.Sprintf("Recovered from panic: %v", r),
						Color:       int32(16760839),
						Timestamp:   new(time.Now()),
					},
				},
			})
			if err != nil {
				welcomer.Logger.Warn().Err(err).Msg("Failed to send webhook message")
			}
		}
	}()

	welcomer.SetupDefaultManagerName(*sandwichManagerName)
	welcomer.SetupLogger(*loggingLevel)
	welcomer.SetupGRPCConnection(
		*sandwichGRPCHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(1024*1024*1024),
		), // Set max message size to 1GB
	)
	welcomer.SetupSandwichClient()
	welcomer.SetupDatabase(ctx, *postgresURL)

	entrypoint(ctx, *webhookUrl, *apply)

	if err := welcomer.Queries.UpsertJobCheckpoint(ctx, database.UpsertJobCheckpointParams{
		JobName:         "check-paypal-memberships",
		LastProcessedTs: time.Now().UTC(),
	}); err != nil {
		welcomer.Logger.Error().Err(err).Msg("Failed to upsert job checkpoint")
	}
	cancel()
}

func entrypoint(ctx context.Context, webhookUrl string, apply bool) {
	paypalClientID, err := requiredEnv("PAYPAL_LIVE_CLIENT_ID")
	if err != nil || paypalClientID == "" {
		welcomer.Logger.Error().Err(err).Msg("missing PAYPAL_LIVE_CLIENT_ID")
		return
	}
	paypalClientSecret, err := requiredEnv("PAYPAL_LIVE_CLIENT_SECRET")
	if err != nil || paypalClientSecret == "" {
		welcomer.Logger.Error().Err(err).Msg("missing PAYPAL_LIVE_CLIENT_SECRET")
		return
	}

	paypalClient, err := paypal.NewClient(paypalClientID, paypalClientSecret, paypal.APIBaseLive)
	if err != nil {
		welcomer.Logger.Error().Err(err).Msg("create PayPal client")
		return
	}

	subscriptions, err := welcomer.Queries.GetAllPaypalSubscriptions(ctx)
	if err != nil {
		welcomer.Logger.Error().Err(err).Msg("get PayPal subscriptions")
		return
	}

	now := time.Now()
	cutoff := now.Add(defaultExpiryWindow)

	mode := "dry-run"
	if apply {
		mode = "apply"
	}

	welcomer.Logger.Info().
		Str("mode", mode).
		Int("subscriptions", len(subscriptions)).
		Dur("window", defaultExpiryWindow).
		Msg("Starting PayPal membership repair")

	var failures []error
	var candidates, updated int

	for _, localSubscription := range subscriptions {
		memberships, err := welcomer.Queries.GetUserMembershipsByTransactionID(ctx, localSubscription.SubscriptionID)
		if err != nil {
			failures = append(failures, fmt.Errorf("get memberships for subscription %s: %w", localSubscription.SubscriptionID, err))
			continue
		}

		var dueMemberships []*database.GetUserMembershipsByTransactionIDRow
		for _, membership := range memberships {
			if database.PlatformType(membership.PlatformType) != database.PlatformTypePaypalSubscription ||
				membership.TransactionID != localSubscription.SubscriptionID ||
				membership.UserID != localSubscription.UserID ||
				!isRenewableMembershipStatus(membership.Status) {
				continue
			}

			expiredStatus := membership.Status == int32(database.MembershipStatusExpired)
			expiredTime := !membership.ExpiresAt.After(now)
			if expiredStatus || expiredTime || !membership.ExpiresAt.After(cutoff) {
				dueMemberships = append(dueMemberships, membership)
			}
		}

		if len(dueMemberships) == 0 {
			continue
		}

		candidates += len(dueMemberships)

		requestCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		paypalSubscription, err := paypalClient.GetSubscriptionDetails(requestCtx, localSubscription.SubscriptionID)
		cancel()
		if err != nil {
			welcomer.Logger.Error().Err(err).
				Str("subscription", localSubscription.SubscriptionID).
				Msg("failed to get live PayPal subscription")

			continue
		}

		if paypalSubscription.SubscriptionStatus != paypal.SubscriptionStatusActive {
			welcomer.Logger.Info().
				Str("subscription", localSubscription.SubscriptionID).
				Str("status", string(paypalSubscription.SubscriptionStatus)).
				Msg("subscription not active; leaving due memberships unchanged")

			continue
		}

		nextBillingAt := time.Time(paypalSubscription.BillingInfo.NextBillingTime)
		if nextBillingAt.IsZero() {
			welcomer.Logger.Error().
				Str("subscription", localSubscription.SubscriptionID).
				Msg("active PayPal subscription has no next billing time")

			continue
		}

		targetExpiry := nextBillingAt.Add(membershipGrace)
		if !targetExpiry.After(now) {
			welcomer.Logger.Warn().
				Str("subscription", localSubscription.SubscriptionID).
				Time("target_expiry", targetExpiry).
				Msg("subscription is active but next billing plus grace is already past; manual review required; no membership or subscription cancelled")

			continue
		}

		for _, membership := range dueMemberships {
			if membership.Status == int32(database.MembershipStatusExpired) || !membership.ExpiresAt.After(now) {
				welcomer.Logger.Warn().
					Str("subscription", localSubscription.SubscriptionID).
					Str("membership", membership.MembershipUuid.String()).
					Str("status", database.MembershipStatus(membership.Status).String()).
					Time("expires_at", membership.ExpiresAt).
					Msg("active subscription has expired membership; it will not be cancelled")
			}

			newExpiry := targetExpiry
			if membership.ExpiresAt.After(newExpiry) {
				newExpiry = membership.ExpiresAt
			}

			if membership.Status == int32(database.MembershipStatusActive) &&
				!newExpiry.After(membership.ExpiresAt) {
				welcomer.Logger.Info().
					Str("subscription", localSubscription.SubscriptionID).
					Str("membership", membership.MembershipUuid.String()).
					Time("expires_at", membership.ExpiresAt).
					Msg("membership already expires at or after target expiry; no update needed")

				continue
			}

			if !apply {
				welcomer.Logger.Info().
					Str("subscription", localSubscription.SubscriptionID).
					Str("membership", membership.MembershipUuid.String()).
					Time("expires_at", membership.ExpiresAt).
					Time("target_expiry", newExpiry).
					Msg("dry-run: would update membership to active with new expiry")

				continue
			}

			err = welcomer.OnPaypalSubscriptionReactivated(ctx, paypal.Subscription{
				SubscriptionDetailResp: *paypalSubscription,
			})
			if err != nil {
				welcomer.Logger.Error().
					Str("subscription", localSubscription.SubscriptionID).
					Str("membership", membership.MembershipUuid.String()).
					Err(err).
					Msg("failed to reactivate PayPal subscription")

				continue
			}

			// rowsAffected, err := welcomer.Queries.UpdateUserMembership(ctx, database.UpdateUserMembershipParams{
			// 	MembershipUuid:  membership.MembershipUuid,
			// 	StartedAt:       membership.StartedAt,
			// 	ExpiresAt:       newExpiry,
			// 	Status:          int32(database.MembershipStatusActive),
			// 	TransactionUuid: membership.TransactionUuid,
			// 	UserID:          membership.UserID,
			// 	GuildID:         membership.GuildID,
			// })
			// if err != nil {
			// 	welcomer.Logger.Error().
			// 		Str("subscription", localSubscription.SubscriptionID).
			// 		Str("membership", membership.MembershipUuid.String()).
			// 		Err(err).
			// 		Msg("failed to update membership")

			// 	continue
			// }
			// if rowsAffected == 0 {
			// 	welcomer.Logger.Warn().
			// 		Str("subscription", localSubscription.SubscriptionID).
			// 		Str("membership", membership.MembershipUuid.String()).
			// 		Msg("membership was not updated")

			// 	continue
			// }

			updated++

			welcomer.Logger.Info().
				Str("subscription", localSubscription.SubscriptionID).
				Str("membership", membership.MembershipUuid.String()).
				Time("expires_at", newExpiry).
				Msg("membership updated successfully")
		}
	}

	welcomer.Logger.Info().
		Str("mode", mode).
		Int("candidates", candidates).
		Int("updated", updated).
		Int("failures", len(failures)).
		Msg("Finished PayPal membership repair")
}

func isRenewableMembershipStatus(status int32) bool {
	return status == int32(database.MembershipStatusActive) ||
		status == int32(database.MembershipStatusIdle) ||
		status == int32(database.MembershipStatusExpired)
}

func requiredEnv(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("required environment variable %s is not set", name)
	}

	return value, nil
}
