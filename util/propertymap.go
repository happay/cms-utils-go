package util

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// ============ Structs =============

type BaseModel struct {
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `sql:"index" json:"deleted_at,omitempty"`
}

// PropertyMap can be used for any non-fixed json to map parsing
type PropertyMap map[string]interface{}

// =========== Exposed (public) Methods - can be called from external packages ============

func (pm PropertyMap) Value() (driver.Value, error) {
	j, err := json.Marshal(pm)
	return j, err
}

func (pm *PropertyMap) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("type assertion .([]byte) failed")
	}

	var i interface{}
	err := json.Unmarshal(source, &i)
	if err != nil {
		return err
	}

	*pm, ok = i.(map[string]interface{})
	if !ok {
		return nil // No need to raise an error, it just means that object is not meant to be recursively checked
	}

	return nil
}

// GetNestedKeyValue will return a nested value using arrays of keys which constitute the path to the nested path
// eg. for nested key, "top/middle/end" => ["top", "middle", "end"] should be passed as keys
// if at any stage a key is not found, it returns empty interface and found as false
//
// NOTE: Its the user responsibility to check if found is true, before using interface value
func GetNestedKeyValue(keys []string, nestedObject interface{}) (value interface{}, found bool) {
	if len(keys) == 0 || nestedObject == nil {
		return
	}
	if len(keys) > 1 {
		switch nestedObject.(type) {
		case map[string]interface{}:
			if nestedSpec, found := nestedObject.(map[string]interface{})[keys[0]]; found {
				return GetNestedKeyValue(keys[1:], nestedSpec)
			}
		case []interface{}:
			for _, nestedSpec := range nestedObject.([]interface{}) {
				if nestedSpec.(string) == keys[0] {
					return GetNestedKeyValue(keys[1:], nestedSpec)
				}
			}
		default:
			return // if any of the above cases doesn't recurse further, it mean key/val was not found
		}
	}
	// base case
	value, found = nestedObject.(map[string]interface{})[keys[0]]
	return
}
