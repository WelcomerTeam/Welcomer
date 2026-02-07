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

func (eg *ErrorGroup) Error(delimiter string) string {
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
