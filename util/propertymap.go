package util

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
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

// ParseEventName will parse the operation id(For each API in Swagger) coming as string and return the event name
func ParseEventName(name string) string {
	if strings.Contains(name, "-") {
		return strings.SplitN(name, "-", 2)[1]
	}
	return name
}

// ReflectInterface will return the underlying concrete value of an interface
func ReflectInterface(data interface{}) interface{} {
	switch data.(type) {
	case map[string]interface{}:
		return data.(map[string]interface{})
	case []interface{}:
		return data.([]interface{})
	default:
		fmt.Printf("unknown datatype %T of given interface %v", data, data)
		return data
	}
}