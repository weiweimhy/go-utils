package jsonutil

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/weiweimhy/go-utils/v6/streamutil"
)

var (
	// ErrEmptyInput is returned when a strict decoder receives no JSON value.
	ErrEmptyInput = errors.New("jsonutil: empty input")
	// ErrTrailingValue is returned when input contains more than one JSON value.
	ErrTrailingValue = errors.New("jsonutil: trailing JSON value")
)

// DuplicateKeyError identifies a repeated object key and its JSON path.
type DuplicateKeyError struct {
	Path string
	Key  string
}

func (e *DuplicateKeyError) Error() string {
	return fmt.Sprintf("jsonutil: duplicate key %q at %s", e.Key, e.Path)
}

// DecodeStrict reads and decodes one JSON value into T. It rejects unknown
// struct fields, duplicate object keys at every nesting level, trailing JSON
// values, and input larger than maxBytes. A non-positive maxBytes disables the
// byte limit.
func DecodeStrict[T any](r io.Reader, maxBytes int64) (T, error) {
	var zero T
	data, err := streamutil.ReadAllLimit(r, maxBytes)
	if err != nil {
		return zero, err
	}
	return UnmarshalStrict[T](data)
}

// UnmarshalStrict decodes one complete JSON value from data with the same
// validation rules as DecodeStrict, except for the byte limit.
func UnmarshalStrict[T any](data []byte) (T, error) {
	var zero T
	if len(bytes.TrimSpace(data)) == 0 {
		return zero, ErrEmptyInput
	}
	if err := validateDocument(data); err != nil {
		return zero, err
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&zero); err != nil {
		return zero, err
	}
	if err := requireEOF(decoder); err != nil {
		return zero, err
	}
	return zero, nil
}

func validateDocument(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := validateValue(decoder, "$"); err != nil {
		if errors.Is(err, io.EOF) {
			return ErrEmptyInput
		}
		return err
	}
	return requireEOF(decoder)
}

func validateValue(decoder *json.Decoder, path string) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}

	switch delim {
	case '{':
		keys := make(map[string]struct{})
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return fmt.Errorf("jsonutil: object key at %s is not a string", path)
			}
			if _, exists := keys[key]; exists {
				return &DuplicateKeyError{Path: path, Key: key}
			}
			keys[key] = struct{}{}
			if err := validateValue(decoder, path+"."+key); err != nil {
				return err
			}
		}
		_, err := decoder.Token()
		return err
	case '[':
		for index := 0; decoder.More(); index++ {
			if err := validateValue(decoder, fmt.Sprintf("%s[%d]", path, index)); err != nil {
				return err
			}
		}
		_, err := decoder.Token()
		return err
	default:
		return fmt.Errorf("jsonutil: unexpected delimiter %q at %s", delim, path)
	}
}

func requireEOF(decoder *json.Decoder) error {
	var extra any
	err := decoder.Decode(&extra)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err == nil {
		return ErrTrailingValue
	}
	return err
}
