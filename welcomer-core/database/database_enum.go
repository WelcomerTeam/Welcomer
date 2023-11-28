// Code generated by go-enum DO NOT EDIT.
// Version:
// Revision:
// Build Date:
// Built By:

package database

import (
	"fmt"
)

const (
	// BackgroundFileTypeUnknown is a BackgroundFileType of type Unknown.
	BackgroundFileTypeUnknown BackgroundFileType = iota
	// BackgroundFileTypeImagePng is a BackgroundFileType of type Image/Png.
	BackgroundFileTypeImagePng
	// BackgroundFileTypeImageJpeg is a BackgroundFileType of type Image/Jpeg.
	BackgroundFileTypeImageJpeg
	// BackgroundFileTypeImageGif is a BackgroundFileType of type Image/Gif.
	BackgroundFileTypeImageGif
	// BackgroundFileTypeImageWebp is a BackgroundFileType of type Image/Webp.
	BackgroundFileTypeImageWebp
)

const _BackgroundFileTypeName = "unknownimage/pngimage/jpegimage/gifimage/webp"

var _BackgroundFileTypeMap = map[BackgroundFileType]string{
	BackgroundFileTypeUnknown:   _BackgroundFileTypeName[0:7],
	BackgroundFileTypeImagePng:  _BackgroundFileTypeName[7:16],
	BackgroundFileTypeImageJpeg: _BackgroundFileTypeName[16:26],
	BackgroundFileTypeImageGif:  _BackgroundFileTypeName[26:35],
	BackgroundFileTypeImageWebp: _BackgroundFileTypeName[35:45],
}

// String implements the Stringer interface.
func (x BackgroundFileType) String() string {
	if str, ok := _BackgroundFileTypeMap[x]; ok {
		return str
	}
	return fmt.Sprintf("BackgroundFileType(%d)", x)
}

var _BackgroundFileTypeValue = map[string]BackgroundFileType{
	_BackgroundFileTypeName[0:7]:   BackgroundFileTypeUnknown,
	_BackgroundFileTypeName[7:16]:  BackgroundFileTypeImagePng,
	_BackgroundFileTypeName[16:26]: BackgroundFileTypeImageJpeg,
	_BackgroundFileTypeName[26:35]: BackgroundFileTypeImageGif,
	_BackgroundFileTypeName[35:45]: BackgroundFileTypeImageWebp,
}

// ParseBackgroundFileType attempts to convert a string to a BackgroundFileType.
func ParseBackgroundFileType(name string) (BackgroundFileType, error) {
	if x, ok := _BackgroundFileTypeValue[name]; ok {
		return x, nil
	}
	return BackgroundFileType(0), fmt.Errorf("%s is not a valid BackgroundFileType", name)
}

// MarshalText implements the text marshaller method.
func (x BackgroundFileType) MarshalText() ([]byte, error) {
	return []byte(x.String()), nil
}

// UnmarshalText implements the text unmarshaller method.
func (x *BackgroundFileType) UnmarshalText(text []byte) error {
	name := string(text)
	tmp, err := ParseBackgroundFileType(name)
	if err != nil {
		return err
	}
	*x = tmp
	return nil
}

const (
	// MembershipStatusUnknown is a MembershipStatus of type Unknown.
	MembershipStatusUnknown MembershipStatus = iota
	// MembershipStatusIdle is a MembershipStatus of type Idle.
	MembershipStatusIdle
	// MembershipStatusActive is a MembershipStatus of type Active.
	MembershipStatusActive
	// MembershipStatusExpired is a MembershipStatus of type Expired.
	MembershipStatusExpired
	// MembershipStatusRefunded is a MembershipStatus of type Refunded.
	MembershipStatusRefunded
	// MembershipStatusRemoved is a MembershipStatus of type Removed.
	MembershipStatusRemoved
)

const _MembershipStatusName = "unknownidleactiveexpiredrefundedremoved"

var _MembershipStatusMap = map[MembershipStatus]string{
	MembershipStatusUnknown:  _MembershipStatusName[0:7],
	MembershipStatusIdle:     _MembershipStatusName[7:11],
	MembershipStatusActive:   _MembershipStatusName[11:17],
	MembershipStatusExpired:  _MembershipStatusName[17:24],
	MembershipStatusRefunded: _MembershipStatusName[24:32],
	MembershipStatusRemoved:  _MembershipStatusName[32:39],
}

// String implements the Stringer interface.
func (x MembershipStatus) String() string {
	if str, ok := _MembershipStatusMap[x]; ok {
		return str
	}
	return fmt.Sprintf("MembershipStatus(%d)", x)
}

var _MembershipStatusValue = map[string]MembershipStatus{
	_MembershipStatusName[0:7]:   MembershipStatusUnknown,
	_MembershipStatusName[7:11]:  MembershipStatusIdle,
	_MembershipStatusName[11:17]: MembershipStatusActive,
	_MembershipStatusName[17:24]: MembershipStatusExpired,
	_MembershipStatusName[24:32]: MembershipStatusRefunded,
	_MembershipStatusName[32:39]: MembershipStatusRemoved,
}

// ParseMembershipStatus attempts to convert a string to a MembershipStatus.
func ParseMembershipStatus(name string) (MembershipStatus, error) {
	if x, ok := _MembershipStatusValue[name]; ok {
		return x, nil
	}
	return MembershipStatus(0), fmt.Errorf("%s is not a valid MembershipStatus", name)
}

// MarshalText implements the text marshaller method.
func (x MembershipStatus) MarshalText() ([]byte, error) {
	return []byte(x.String()), nil
}

// UnmarshalText implements the text unmarshaller method.
func (x *MembershipStatus) UnmarshalText(text []byte) error {
	name := string(text)
	tmp, err := ParseMembershipStatus(name)
	if err != nil {
		return err
	}
	*x = tmp
	return nil
}

const (
	// MembershipTypeUnknown is a MembershipType of type Unknown.
	MembershipTypeUnknown MembershipType = iota
	// MembershipTypeLegacyCustomBackgrounds is a MembershipType of type LegacyCustomBackgrounds.
	MembershipTypeLegacyCustomBackgrounds
	// MembershipTypeLegacyWelcomerPro1 is a MembershipType of type LegacyWelcomerPro1.
	MembershipTypeLegacyWelcomerPro1
	// MembershipTypeLegacyWelcomerPro3 is a MembershipType of type LegacyWelcomerPro3.
	MembershipTypeLegacyWelcomerPro3
	// MembershipTypeLegacyWelcomerPro5 is a MembershipType of type LegacyWelcomerPro5.
	MembershipTypeLegacyWelcomerPro5
	// MembershipTypeWelcomerPro is a MembershipType of type WelcomerPro.
	MembershipTypeWelcomerPro
	// MembershipTypeCustomBackgrounds is a MembershipType of type CustomBackgrounds.
	MembershipTypeCustomBackgrounds
)

const _MembershipTypeName = "unknownlegacyCustomBackgroundslegacyWelcomerPro1legacyWelcomerPro3legacyWelcomerPro5welcomerProcustomBackgrounds"

var _MembershipTypeMap = map[MembershipType]string{
	MembershipTypeUnknown:                 _MembershipTypeName[0:7],
	MembershipTypeLegacyCustomBackgrounds: _MembershipTypeName[7:30],
	MembershipTypeLegacyWelcomerPro1:      _MembershipTypeName[30:48],
	MembershipTypeLegacyWelcomerPro3:      _MembershipTypeName[48:66],
	MembershipTypeLegacyWelcomerPro5:      _MembershipTypeName[66:84],
	MembershipTypeWelcomerPro:             _MembershipTypeName[84:95],
	MembershipTypeCustomBackgrounds:       _MembershipTypeName[95:112],
}

// String implements the Stringer interface.
func (x MembershipType) String() string {
	if str, ok := _MembershipTypeMap[x]; ok {
		return str
	}
	return fmt.Sprintf("MembershipType(%d)", x)
}

var _MembershipTypeValue = map[string]MembershipType{
	_MembershipTypeName[0:7]:    MembershipTypeUnknown,
	_MembershipTypeName[7:30]:   MembershipTypeLegacyCustomBackgrounds,
	_MembershipTypeName[30:48]:  MembershipTypeLegacyWelcomerPro1,
	_MembershipTypeName[48:66]:  MembershipTypeLegacyWelcomerPro3,
	_MembershipTypeName[66:84]:  MembershipTypeLegacyWelcomerPro5,
	_MembershipTypeName[84:95]:  MembershipTypeWelcomerPro,
	_MembershipTypeName[95:112]: MembershipTypeCustomBackgrounds,
}

// ParseMembershipType attempts to convert a string to a MembershipType.
func ParseMembershipType(name string) (MembershipType, error) {
	if x, ok := _MembershipTypeValue[name]; ok {
		return x, nil
	}
	return MembershipType(0), fmt.Errorf("%s is not a valid MembershipType", name)
}

// MarshalText implements the text marshaller method.
func (x MembershipType) MarshalText() ([]byte, error) {
	return []byte(x.String()), nil
}

// UnmarshalText implements the text unmarshaller method.
func (x *MembershipType) UnmarshalText(text []byte) error {
	name := string(text)
	tmp, err := ParseMembershipType(name)
	if err != nil {
		return err
	}
	*x = tmp
	return nil
}

const (
	// PlatformTypeUnknown is a PlatformType of type Unknown.
	PlatformTypeUnknown PlatformType = iota
	// PlatformTypePaypal is a PlatformType of type Paypal.
	PlatformTypePaypal
	// PlatformTypePatreon is a PlatformType of type Patreon.
	PlatformTypePatreon
	// PlatformTypeStripe is a PlatformType of type Stripe.
	PlatformTypeStripe
)

const _PlatformTypeName = "unknownpaypalpatreonstripe"

var _PlatformTypeMap = map[PlatformType]string{
	PlatformTypeUnknown: _PlatformTypeName[0:7],
	PlatformTypePaypal:  _PlatformTypeName[7:13],
	PlatformTypePatreon: _PlatformTypeName[13:20],
	PlatformTypeStripe:  _PlatformTypeName[20:26],
}

// String implements the Stringer interface.
func (x PlatformType) String() string {
	if str, ok := _PlatformTypeMap[x]; ok {
		return str
	}
	return fmt.Sprintf("PlatformType(%d)", x)
}

var _PlatformTypeValue = map[string]PlatformType{
	_PlatformTypeName[0:7]:   PlatformTypeUnknown,
	_PlatformTypeName[7:13]:  PlatformTypePaypal,
	_PlatformTypeName[13:20]: PlatformTypePatreon,
	_PlatformTypeName[20:26]: PlatformTypeStripe,
}

// ParsePlatformType attempts to convert a string to a PlatformType.
func ParsePlatformType(name string) (PlatformType, error) {
	if x, ok := _PlatformTypeValue[name]; ok {
		return x, nil
	}
	return PlatformType(0), fmt.Errorf("%s is not a valid PlatformType", name)
}

// MarshalText implements the text marshaller method.
func (x PlatformType) MarshalText() ([]byte, error) {
	return []byte(x.String()), nil
}

// UnmarshalText implements the text unmarshaller method.
func (x *PlatformType) UnmarshalText(text []byte) error {
	name := string(text)
	tmp, err := ParsePlatformType(name)
	if err != nil {
		return err
	}
	*x = tmp
	return nil
}

const (
	// ScienceEventTypeUnknown is a ScienceEventType of type Unknown.
	ScienceEventTypeUnknown ScienceEventType = iota
	// ScienceEventTypeGuildJoin is a ScienceEventType of type GuildJoin.
	ScienceEventTypeGuildJoin
	// ScienceEventTypeGuildLeave is a ScienceEventType of type GuildLeave.
	ScienceEventTypeGuildLeave
)

const _ScienceEventTypeName = "unknownguildJoinguildLeave"

var _ScienceEventTypeMap = map[ScienceEventType]string{
	ScienceEventTypeUnknown:    _ScienceEventTypeName[0:7],
	ScienceEventTypeGuildJoin:  _ScienceEventTypeName[7:16],
	ScienceEventTypeGuildLeave: _ScienceEventTypeName[16:26],
}

// String implements the Stringer interface.
func (x ScienceEventType) String() string {
	if str, ok := _ScienceEventTypeMap[x]; ok {
		return str
	}
	return fmt.Sprintf("ScienceEventType(%d)", x)
}

var _ScienceEventTypeValue = map[string]ScienceEventType{
	_ScienceEventTypeName[0:7]:   ScienceEventTypeUnknown,
	_ScienceEventTypeName[7:16]:  ScienceEventTypeGuildJoin,
	_ScienceEventTypeName[16:26]: ScienceEventTypeGuildLeave,
}

// ParseScienceEventType attempts to convert a string to a ScienceEventType.
func ParseScienceEventType(name string) (ScienceEventType, error) {
	if x, ok := _ScienceEventTypeValue[name]; ok {
		return x, nil
	}
	return ScienceEventType(0), fmt.Errorf("%s is not a valid ScienceEventType", name)
}

// MarshalText implements the text marshaller method.
func (x ScienceEventType) MarshalText() ([]byte, error) {
	return []byte(x.String()), nil
}

// UnmarshalText implements the text unmarshaller method.
func (x *ScienceEventType) UnmarshalText(text []byte) error {
	name := string(text)
	tmp, err := ParseScienceEventType(name)
	if err != nil {
		return err
	}
	*x = tmp
	return nil
}

const (
	// ScienceGuildEventTypeUnknown is a ScienceGuildEventType of type Unknown.
	ScienceGuildEventTypeUnknown ScienceGuildEventType = iota
	// ScienceGuildEventTypeJoin is a ScienceGuildEventType of type Join.
	ScienceGuildEventTypeJoin
	// ScienceGuildEventTypeLeave is a ScienceGuildEventType of type Leave.
	ScienceGuildEventTypeLeave
	// ScienceGuildEventTypeUserJoin is a ScienceGuildEventType of type UserJoin.
	ScienceGuildEventTypeUserJoin
	// ScienceGuildEventTypeUserLeave is a ScienceGuildEventType of type UserLeave.
	ScienceGuildEventTypeUserLeave
	// ScienceGuildEventTypeUserWelcomed is a ScienceGuildEventType of type UserWelcomed.
	ScienceGuildEventTypeUserWelcomed
	// ScienceGuildEventTypeTimeRoleGiven is a ScienceGuildEventType of type TimeRoleGiven.
	ScienceGuildEventTypeTimeRoleGiven
	// ScienceGuildEventTypeBorderwallChallenge is a ScienceGuildEventType of type BorderwallChallenge.
	ScienceGuildEventTypeBorderwallChallenge
	// ScienceGuildEventTypeBorderwallCompleted is a ScienceGuildEventType of type BorderwallCompleted.
	ScienceGuildEventTypeBorderwallCompleted
	// ScienceGuildEventTypeTempChannelCreated is a ScienceGuildEventType of type TempChannelCreated.
	ScienceGuildEventTypeTempChannelCreated
	// ScienceGuildEventTypeMembershipReceived is a ScienceGuildEventType of type MembershipReceived.
	ScienceGuildEventTypeMembershipReceived
	// ScienceGuildEventTypeMembershipRemoved is a ScienceGuildEventType of type MembershipRemoved.
	ScienceGuildEventTypeMembershipRemoved
)

const _ScienceGuildEventTypeName = "unknownjoinleaveuserJoinuserLeaveuserWelcomedtimeRoleGivenborderwallChallengeborderwallCompletedtempChannelCreatedmembershipReceivedmembershipRemoved"

var _ScienceGuildEventTypeMap = map[ScienceGuildEventType]string{
	ScienceGuildEventTypeUnknown:             _ScienceGuildEventTypeName[0:7],
	ScienceGuildEventTypeJoin:                _ScienceGuildEventTypeName[7:11],
	ScienceGuildEventTypeLeave:               _ScienceGuildEventTypeName[11:16],
	ScienceGuildEventTypeUserJoin:            _ScienceGuildEventTypeName[16:24],
	ScienceGuildEventTypeUserLeave:           _ScienceGuildEventTypeName[24:33],
	ScienceGuildEventTypeUserWelcomed:        _ScienceGuildEventTypeName[33:45],
	ScienceGuildEventTypeTimeRoleGiven:       _ScienceGuildEventTypeName[45:58],
	ScienceGuildEventTypeBorderwallChallenge: _ScienceGuildEventTypeName[58:77],
	ScienceGuildEventTypeBorderwallCompleted: _ScienceGuildEventTypeName[77:96],
	ScienceGuildEventTypeTempChannelCreated:  _ScienceGuildEventTypeName[96:114],
	ScienceGuildEventTypeMembershipReceived:  _ScienceGuildEventTypeName[114:132],
	ScienceGuildEventTypeMembershipRemoved:   _ScienceGuildEventTypeName[132:149],
}

// String implements the Stringer interface.
func (x ScienceGuildEventType) String() string {
	if str, ok := _ScienceGuildEventTypeMap[x]; ok {
		return str
	}
	return fmt.Sprintf("ScienceGuildEventType(%d)", x)
}

var _ScienceGuildEventTypeValue = map[string]ScienceGuildEventType{
	_ScienceGuildEventTypeName[0:7]:     ScienceGuildEventTypeUnknown,
	_ScienceGuildEventTypeName[7:11]:    ScienceGuildEventTypeJoin,
	_ScienceGuildEventTypeName[11:16]:   ScienceGuildEventTypeLeave,
	_ScienceGuildEventTypeName[16:24]:   ScienceGuildEventTypeUserJoin,
	_ScienceGuildEventTypeName[24:33]:   ScienceGuildEventTypeUserLeave,
	_ScienceGuildEventTypeName[33:45]:   ScienceGuildEventTypeUserWelcomed,
	_ScienceGuildEventTypeName[45:58]:   ScienceGuildEventTypeTimeRoleGiven,
	_ScienceGuildEventTypeName[58:77]:   ScienceGuildEventTypeBorderwallChallenge,
	_ScienceGuildEventTypeName[77:96]:   ScienceGuildEventTypeBorderwallCompleted,
	_ScienceGuildEventTypeName[96:114]:  ScienceGuildEventTypeTempChannelCreated,
	_ScienceGuildEventTypeName[114:132]: ScienceGuildEventTypeMembershipReceived,
	_ScienceGuildEventTypeName[132:149]: ScienceGuildEventTypeMembershipRemoved,
}

// ParseScienceGuildEventType attempts to convert a string to a ScienceGuildEventType.
func ParseScienceGuildEventType(name string) (ScienceGuildEventType, error) {
	if x, ok := _ScienceGuildEventTypeValue[name]; ok {
		return x, nil
	}
	return ScienceGuildEventType(0), fmt.Errorf("%s is not a valid ScienceGuildEventType", name)
}

// MarshalText implements the text marshaller method.
func (x ScienceGuildEventType) MarshalText() ([]byte, error) {
	return []byte(x.String()), nil
}

// UnmarshalText implements the text unmarshaller method.
func (x *ScienceGuildEventType) UnmarshalText(text []byte) error {
	name := string(text)
	tmp, err := ParseScienceGuildEventType(name)
	if err != nil {
		return err
	}
	*x = tmp
	return nil
}

const (
	// TransactionStatusUnknown is a TransactionStatus of type Unknown.
	TransactionStatusUnknown TransactionStatus = iota
	// TransactionStatusPending is a TransactionStatus of type Pending.
	TransactionStatusPending
	// TransactionStatusCompleted is a TransactionStatus of type Completed.
	TransactionStatusCompleted
	// TransactionStatusRefunded is a TransactionStatus of type Refunded.
	TransactionStatusRefunded
)

const _TransactionStatusName = "unknownpendingcompletedrefunded"

var _TransactionStatusMap = map[TransactionStatus]string{
	TransactionStatusUnknown:   _TransactionStatusName[0:7],
	TransactionStatusPending:   _TransactionStatusName[7:14],
	TransactionStatusCompleted: _TransactionStatusName[14:23],
	TransactionStatusRefunded:  _TransactionStatusName[23:31],
}

// String implements the Stringer interface.
func (x TransactionStatus) String() string {
	if str, ok := _TransactionStatusMap[x]; ok {
		return str
	}
	return fmt.Sprintf("TransactionStatus(%d)", x)
}

var _TransactionStatusValue = map[string]TransactionStatus{
	_TransactionStatusName[0:7]:   TransactionStatusUnknown,
	_TransactionStatusName[7:14]:  TransactionStatusPending,
	_TransactionStatusName[14:23]: TransactionStatusCompleted,
	_TransactionStatusName[23:31]: TransactionStatusRefunded,
}

// ParseTransactionStatus attempts to convert a string to a TransactionStatus.
func ParseTransactionStatus(name string) (TransactionStatus, error) {
	if x, ok := _TransactionStatusValue[name]; ok {
		return x, nil
	}
	return TransactionStatus(0), fmt.Errorf("%s is not a valid TransactionStatus", name)
}

// MarshalText implements the text marshaller method.
func (x TransactionStatus) MarshalText() ([]byte, error) {
	return []byte(x.String()), nil
}

// UnmarshalText implements the text unmarshaller method.
func (x *TransactionStatus) UnmarshalText(text []byte) error {
	name := string(text)
	tmp, err := ParseTransactionStatus(name)
	if err != nil {
		return err
	}
	*x = tmp
	return nil
}
