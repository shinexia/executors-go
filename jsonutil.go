package executors

import (
	"encoding/json"
	"fmt"
)

// ToJSON marshal an object to JSON string, use fmt when marshal failed
func ToJSON(obj any) string {
	data, err := json.Marshal(obj)
	if err != nil {
		return fmt.Sprintf("%v", obj)
	}
	return string(data)
}

// ToJSON marshal an object to JSON string, use fmt when marshal failed
func ToPrettyJSON(obj any) string {
	data, err := json.MarshalIndent(obj, "  ", "  ")
	if err != nil {
		return fmt.Sprintf("%v", obj)
	}
	return string(data)
}

func ToLogField(obj any) any {
	data, err := json.Marshal(obj)
	if err != nil {
		return obj
	}
	return json.RawMessage(data)
}
