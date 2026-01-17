package welcomer

//go:generate go-enum -f=$GOFILE --marshal

// ENUM(create, edit)
type IngestMessageEventType int16

// ENUM(join, leave, checkpoint)
type IngestVoiceChannelEventType int16
