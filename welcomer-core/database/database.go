package database

//go:generate go-enum -f=$GOFILE --marshal

// ENUM(unknown)
type ScienceEventType int32

// ENUM(unknown, userJoin, userLeave, userWelcomed, timeRoleGiven, borderwallChallenge, borderwallCompleted, tempChannelCreated, membershipReceived, membershipRemoved, guildJoin, guildLeave, guildOnboarded, guildUserOnboarded, welcomeMessageRemoved, userLeftMessage, leaverMessageRemoved)
type ScienceGuildEventType int32

// ENUM(unknown, idle, active, expired, refunded, removed)
type MembershipStatus int32

// ENUM(unknown, legacyCustomBackgrounds, legacyWelcomerPro, welcomerPro, customBackgrounds)
type MembershipType int32

// ENUM(unknown, paypal, patreon, stripe, paypal_subscription, discord)
type PlatformType int32

// ENUM(unknown, pending, completed, refunded)
type TransactionStatus int32

// ENUM(default, commas, dots, indian, arabic)
type NumberLocale int32
