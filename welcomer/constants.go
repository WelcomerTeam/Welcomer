package welcomer

type PunishmentType int16

const (
	PunishmentTypeUnknown PunishmentType = iota
	PunishmentTypeWarn
	PunishmentTypeKick
	PunishmentTypeTimeout
	PunishmentTypeBan
	PunishmentTypeUnban
	PunishmentTypeSoftBan
)

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
	ScienceGuildEventTypeMigration
	ScienceGuildEventTypeTimeRoleGiven
	ScienceGuildEventTypeBorderwallChallenge
	ScienceGuildEventTypeCompleted
	ScienceGuildEventTypeUserJoin
	ScienceGuildEventTypeUserLeave
	ScienceGuildEventTypeUserWelcomed
)

type MembershipStatus int64

const (
	MembershipStatusUnknown MembershipStatus = iota
	MembershipStatusIdle
	MembershipStatusActive
	MembershipStatusExpired
	MembershipStatusRefunded
	MembershipStatusRemoved
	MembershipStatusTransfered
)

type MembershipType int64

const (
	MembershipTypeUnknown MembershipType = iota
	MembershipTypeLegacyCustomBackgrounds
	MembershipTypeLegacyWelcomerPro1
	MembershipTypeLegacyWelcomerPro3
	MembershipTypeLegacyWelcomerPro5
	MembershipTypeWelcomerPro1
	MembershipTypeWelcomerPro3
	MembershipTypeWelcomerPro5
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
	TransactionStatusUnclaimed
)
