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

// ConvertMapToStruct, will convert the map into struct passed in type parameter([T any]).
func ConvertMapToStruct[T any](mapData map[string]interface{}) (resp T, err error) {
	// Convert map to json string
	jsonStr, err := json.Marshal(mapData)
	if err != nil {
		return
	}
	// Convert json string to struct
	err = json.Unmarshal(jsonStr, &resp)
	return
}
