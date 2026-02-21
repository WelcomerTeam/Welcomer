package welcomer

import "strings"

type ErrorGroup []error

func NewErrorGroup() *ErrorGroup {
	return &ErrorGroup{}
}

func (eg *ErrorGroup) Empty() bool {
	return len(*eg) == 0
}

func (eg *ErrorGroup) Add(err error) {
	*eg = append(*eg, err)
}

func (eg *ErrorGroup) Error() string {
	return eg.ErrorWithDelimiter("; ")
}

func (eg *ErrorGroup) AsStandardError() error {
	if eg.Empty() {
		return nil
	}

	return ErrorGroupError(eg.Error())
}

func (eg *ErrorGroup) ErrorWithDelimiter(delimiter string) string {
	if eg.Empty() {
		return ""
	}

	var errorBuilder strings.Builder

	for i, err := range *eg {
		errorBuilder.WriteString(err.Error())

		if i < len(*eg)-1 {
			errorBuilder.WriteString(delimiter)
		}
	}

	return errorBuilder.String()
}

// Hack to allow for err != nil checks
type ErrorGroupError string

func (e ErrorGroupError) Error() string {
	return string(e)
}
