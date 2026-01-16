package welcomer

//go:generate go-enum -f=$GOFILE --marshal

// ENUM(create, edit)
type IngestMessageEventType int16

// ENUM(join, leave)
type IngestVoiceChannelEventType int16
