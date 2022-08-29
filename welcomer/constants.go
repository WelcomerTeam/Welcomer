package welcomer

type ScienceEventType int64

const (
	ScienceEventTypeUnknown ScienceEventType = iota
	ScienceEventTypeGuildJoin
	ScienceEventTypeGuildLeave
)

type ScienceGuildEventType int64

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

type MembershipStatus int64

const (
	MembershipStatusUnknown MembershipStatus = iota
	MembershipStatusIdle
	MembershipStatusActive
	MembershipStatusExpired
	MembershipStatusRefunded
	MembershipStatusRemoved
)

type MembershipType int64

const (
	MembershipTypeUnknown MembershipType = iota
	MembershipTypeLegacyCustomBackgrounds
	MembershipTypeLegacyWelcomerPro1
	MembershipTypeLegacyWelcomerPro3
	MembershipTypeLegacyWelcomerPro5
	MembershipTypeWelcomerPro
	MembershipTypeCustomBackgrounds
)

type PlatformType int64

const (
	PlatformTypeUnknown PlatformType = iota
	PlatformTypePaypal
	PlatformTypePatreon
	PlatformTypeStripe
)

type TransactionStatus int64

const (
	TransactionStatusUnknown TransactionStatus = iota
	TransactionStatusPending
	TransactionStatusCompleted
	TransactionStatusRefunded
)
