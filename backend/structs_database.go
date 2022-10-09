package backend

import (
	"strconv"

	"github.com/jackc/pgtype"
	gotils_strconv "github.com/savsgio/gotils/strconv"
)

func Int64ToStringPointer(value int64) *string {
	if value == 0 {
		return nil
	}

	v := strconv.FormatInt(value, int64Base)
	return &v
}

func StringPointerToInt64(value *string) int64 {
	if value == nil {
		return 0
	}

	v, _ := strconv.ParseInt(*value, int64Base, int64BitSize)
	return v
}

func StringToJSONB(value string) pgtype.JSONB {
	v := pgtype.JSONB{}

	if value == "" {
		value = `{}`
	}

	err := v.Set(gotils_strconv.S2B(value))
	if err != nil {
		_ = v.Set(`{}`)
	}

	return v
}

func JSONBToString(value pgtype.JSONB) string {
	return gotils_strconv.B2S(value.Bytes)
}
