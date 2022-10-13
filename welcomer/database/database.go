package database

//go:generate go-enum -f=$GOFILE --marshal

// ENUM(unknown, guildJoin, guildLeave)
type ScienceEventType int32

// ENUM(unknown, join, leave, userJoin, userLeave, userWelcomed, timeRoleGiven, borderwallChallenge, borderwallCompleted, tempChannelCreated, membershipReceived, membershipRemoved)
type ScienceGuildEventType int32

// ENUM(unknown, idle, active, expired, refunded, removed)
type MembershipStatus int32

// ENUM(unknown, legacyCustomBackgrounds, legacyWelcomerPro1, legacyWelcomerPro3, legacyWelcomerPro5, welcomerPro, customBackgrounds)
type MembershipType int32

// ENUM(unknown, paypal, patreon, stripe)
type PlatformType int32

// ENUM(unknown, pending, completed, refunded)
type TransactionStatus int32

// ENUM(unknown, PNG, JPG, GIF, WEBP)
type BackgroundFileType int32
