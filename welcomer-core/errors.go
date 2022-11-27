package welcomer

import "fmt"

var (
	ErrInvalidJSON       = fmt.Errorf("invalid json")
	ErrInvalidColour     = fmt.Errorf("colour format is not recognised")
	ErrInvalidBackground = fmt.Errorf("invalid background")
)
