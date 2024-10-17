package dto

import "encoding/json"

func Unmarshal(data []byte, v any) error {
	if v == nil {
		return nil
	} else if len(data) == 0 || data == nil {
		v = nil
		return nil
	} else {
		return json.Unmarshal(data, v)
	}
}

func Marshal(v any) ([]byte, error) {
	if v == nil {
		return nil, nil
	}
	return json.Marshal(v)
}
