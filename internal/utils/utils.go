package utils

import (
	"encoding/json"
	"io"
)

func ReadJSON(r io.ReadCloser, data interface{}) error {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(data)
	if err != nil {
		return err
	}

	return nil
}
