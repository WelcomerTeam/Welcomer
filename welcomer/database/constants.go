package database

type ScienceEventType int32

const (
	ScienceEventTypeUnknown ScienceEventType = iota
	ScienceEventTypeGuildJoin
	ScienceEventTypeGuildLeave
)

type ScienceGuildEventType int32

const (
	ScienceGuildEventTypeUnknown ScienceGuildEventType = iota

	ScienceGuildEventTypeJoin
	ScienceGuildEventTypeLeave

	ScienceGuildEventTypeUserJoin
	ScienceGuildEventTypeUserLeave
	ScienceGuildEventTypeUserWelcomed

	ScienceGuildEventTypeTimeRoleGiven

	ScienceGuildEventTypeBorderwallChallenge
	ScienceGuildEventTypeBorderwallCompleted

	ScienceGuildEventTypeTempChannelCreated

	ScienceGuildEventTypeMembershipReceived
	ScienceGuildEventTypeMembershipRemoved
)

type MembershipStatus int32

const (
	MembershipStatusUnknown MembershipStatus = iota
	MembershipStatusIdle
	MembershipStatusActive
	MembershipStatusExpired
	MembershipStatusRefunded
	MembershipStatusRemoved
)

type MembershipType int32

const (
	MembershipTypeUnknown MembershipType = iota
	MembershipTypeLegacyCustomBackgrounds
	MembershipTypeLegacyWelcomerPro1
	MembershipTypeLegacyWelcomerPro3
	MembershipTypeLegacyWelcomerPro5
	MembershipTypeWelcomerPro
	MembershipTypeCustomBackgrounds
)

type PlatformType int32

const (
	PlatformTypeUnknown PlatformType = iota
	PlatformTypePaypal
	PlatformTypePatreon
	PlatformTypeStripe
)

type TransactionStatus int32

const (
	TransactionStatusUnknown TransactionStatus = iota
	TransactionStatusPending
	TransactionStatusCompleted
	TransactionStatusRefunded
)
