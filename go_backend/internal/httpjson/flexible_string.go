package httpjson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type FlexibleString string

func (f *FlexibleString) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if bytes.Equal(trimmed, []byte("null")) {
		*f = ""
		return nil
	}

	var asString string
	if err := json.Unmarshal(trimmed, &asString); err == nil {
		*f = FlexibleString(strings.TrimSpace(asString))
		return nil
	}

	var asNumber json.Number
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.UseNumber()
	if err := decoder.Decode(&asNumber); err == nil {
		*f = FlexibleString(asNumber.String())
		return nil
	}

	return fmt.Errorf("must be a string or number")
}

func (f FlexibleString) String() string {
	return string(f)
}
