package welcomer

import (
	"testing"
)

type testStruct struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Extra string // no tag
	Skip  string `json:"-"`
	Inner struct {
		Field int
	}
	privateField string // unexported
}

func TestCompareStructs_Basic(t *testing.T) {
	oldStruct := testStruct{
		ID:    1,
		Name:  "Alice",
		Age:   20,
		Extra: "foo",
		Skip:  "should be skipped",
		Inner: struct{ Field int }{Field: 10},
	}

	newStruct := testStruct{
		ID:    2,
		Name:  "Bob",
		Age:   20,
		Extra: "bar",
		Skip:  "should be skipped",
		Inner: struct{ Field int }{Field: 20},
	}

	got, _ := CompareStructs(oldStruct, newStruct)

	want := CompareStructResult{
		"id":    [2]any{1, 2},
		"name":  [2]any{"Alice", "Bob"},
		"Extra": [2]any{"foo", "bar"},
		"Inner": [2]any{struct{ Field int }{Field: 10}, struct{ Field int }{Field: 20}},
	}

	if len(got) != len(want) {
		t.Errorf("expected %d differences, got %d", len(want), len(got))
	}

	for k, v := range want {
		if diff, ok := got[k]; !ok || diff != v {
			t.Errorf("field %q: expected %v, got %v", k, v, got[k])
		}
	}

	// Ensure skipped fields are not present
	if _, ok := got["Skip"]; ok {
		t.Errorf("field 'Skip' should be skipped")
	}

	if _, ok := got["privateField"]; ok {
		t.Errorf("unexported field 'privateField' should be skipped")
	}
}

func TestCompareStructs_NoDifferences(t *testing.T) {
	t.Parallel()

	oldStruct := testStruct{ID: 1, Name: "A", Age: 10, Extra: "x"}

	newStruct := testStruct{ID: 1, Name: "A", Age: 10, Extra: "x"}

	got, _ := CompareStructs(oldStruct, newStruct)
	if len(got) != 0 {
		t.Errorf("expected no differences, got %v", got)
	}
}

func TestCompareStructs_JSONTagWithComma(t *testing.T) {
	t.Parallel()

	type tagStruct struct {
		Field1 int `json:"field1,omitempty"`
		Field2 int `json:"field2"`
	}

	oldStruct := tagStruct{Field1: 1, Field2: 2}

	newStruct := tagStruct{Field1: 2, Field2: 2}

	got, _ := CompareStructs(oldStruct, newStruct)

	if _, ok := got["field1"]; !ok {
		t.Errorf("expected field 'field1' to be present")
	}

	if _, ok := got["field2"]; ok {
		t.Errorf("expected field 'field2' to be absent (no change)")
	}
}

func TestCompareStructs_EmptyStruct(t *testing.T) {
	t.Parallel()

	type emptyStruct struct{}

	oldStruct := emptyStruct{}

	newStruct := emptyStruct{}

	got, _ := CompareStructs(oldStruct, newStruct)
	if len(got) != 0 {
		t.Errorf("expected no differences for empty struct, got %v", got)
	}
}
