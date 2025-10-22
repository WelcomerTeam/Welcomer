package welcomer

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type CompareStructResult map[string][2]any

func CompareStructsAsJSON[T comparable](oldStruct, newStruct T) ([]byte, bool, error) {
	compareResults, hasChanges := CompareStructs(oldStruct, newStruct)

	jsonData, err := json.Marshal(compareResults)
	if err != nil {
		return nil, false, fmt.Errorf("failed to marshal compareStructs: %w", err)
	}

	return jsonData, hasChanges, nil
}

func CompareStructs[T comparable](oldStruct, newStruct T) (CompareStructResult, bool) {
	result := CompareStructResult{}

	oldValue := reflect.ValueOf(oldStruct)
	newValue := reflect.ValueOf(newStruct)

	t := oldValue.Type()

	for fieldIndex := range t.NumField() {
		field := t.Field(fieldIndex)

		// Skip nested structs
		if field.Type.Kind() == reflect.Struct {
			continue
		}

		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		// Use JSON tag or fallback to field name
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = field.Name
		} else if jsonTag == "-" {
			// Skip fields that do not appear
			continue
		} else if comma := findComma(jsonTag); comma != -1 {
			jsonTag = jsonTag[:comma]
		}

		oldField := oldValue.Field(fieldIndex).Interface()
		newField := newValue.Field(fieldIndex).Interface()

		if !reflect.DeepEqual(oldField, newField) {
			result[jsonTag] = [2]any{oldField, newField}
		}
	}

	return result, len(result) > 0
}

func findComma(s string) int {
	for i, ch := range s {
		if ch == ',' {
			return i
		}
	}

	return -1
}
