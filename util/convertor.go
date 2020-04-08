package util

import (
	"encoding/json"
	"time"
)

const DobFormatLayout = "2006-01-02"

// ParseDateStr Convert Date String (YYYY-MM-DD) to time.Time object
func ParseDateStr(dateString string) (time.Time, error) {
	return time.Parse(DobFormatLayout, dateString)
}

// MapToStruct Convert Map to Specified Struct
func MapToStruct(sourceMap map[string]interface{}, targetStruct interface{}) error {
	b, err := json.Marshal(sourceMap)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, targetStruct); err != nil {
		return err
	}

	return nil
}
