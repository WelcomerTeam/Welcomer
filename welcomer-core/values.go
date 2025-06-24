package welcomer

import (
	"github.com/gofrs/uuid"
)

// Contains the structure for the "values" attribute in sandwich configuration files.

type ApplicationValues struct {
	IsCustomBot bool      `json:"custom_bot"`
	BotUUID     uuid.UUID `json:"bot_uuid"`
}
