package httpjson

import (
	"encoding/json"
	"errors"
	"io"
)

func DecodeStrict(body io.Reader, target any) error {
	if body == nil {
		return errors.New("request body is required")
	}

	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func DecodeStrictAllowEOF(body io.Reader, target any) error {
	if body == nil {
		return nil
	}

	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}
