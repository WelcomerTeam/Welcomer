package backend

type ModerationCheckupStatus uint8

const (
	ModerationCheckupStatusUnknown ModerationCheckupStatus = iota
	ModerationCheckupStatusPending
	ModerationCheckupStatusApproved
	ModerationCheckupStatusRejected
)
