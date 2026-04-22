package welcomer

import "encoding/json"

//go:generate go-enum -f=$GOFILE --marshal

// ENUM(never, only_additions, always)
type PollResubmissionOption string

// ENUM(always, after_voting, after_end, never)
type PollResultVisibilityOption string

func UnmarshalAnswersListJSON(answersJSON []byte) (answers []string) {
	_ = json.Unmarshal(answersJSON, &answers)

	return
}

func MarshalAnswersListJSON(answers []string) (answersJSON []byte) {
	answersJSON, _ = json.Marshal(answers)

	return
}
