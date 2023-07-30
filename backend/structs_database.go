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

func BytesToJSONB(value []byte) pgtype.JSONB {
	v := pgtype.JSONB{}

	if len(value) == 0 {
		value = []byte{123, 125} // {}
	}

	err := v.Set(value)
	if err != nil {
		_ = v.Set([]byte{123, 125}) // {}
	}

	return v
}

func JSONBToBytes(value pgtype.JSONB) []byte {
	return value.Bytes
}

func StringToJSONB(value string) pgtype.JSONB {
	return BytesToJSONB(gotils_strconv.S2B(value))
}

func JSONBToString(value pgtype.JSONB) string {
	return gotils_strconv.B2S(JSONBToBytes(value))
}
