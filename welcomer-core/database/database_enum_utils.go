package database

func (x MembershipStatus) Label() string {
	switch x {
	case MembershipStatusUnknown:
		return "Unknown"
	case MembershipStatusIdle:
		return "Idle"
	case MembershipStatusActive:
		return "Active"
	case MembershipStatusExpired:
		return "Expired"
	case MembershipStatusRefunded:
		return "Refunded"
	case MembershipStatusRemoved:
		return "Removed"
	}

	return "Unknown"
}

func (x MembershipType) Label() string {
	switch x {
	case MembershipTypeUnknown:
		return "Unknown"
	case MembershipTypeLegacyCustomBackgrounds:
		return "Legacy Custom Backgrounds"
	case MembershipTypeLegacyWelcomerPro:
		return "Legacy Welcomer Pro"
	case MembershipTypeWelcomerPro:
		return "Welcomer Pro"
	case MembershipTypeCustomBackgrounds:
		return "Custom Backgrounds"
	}

	return "Unknown"
}
