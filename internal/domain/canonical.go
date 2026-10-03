// Package domain contains framework-neutral records and deterministic primitives.
package domain

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"unicode/utf8"
)

// CanonicalJSONError reports input outside the rwt-cjson-1 value vocabulary.
type CanonicalJSONError struct {
	Reason string
	Cause  error
}

// Error describes the rejection without echoing the input document.
func (e *CanonicalJSONError) Error() string {
	return "invalid rwt-cjson-1 JSON: " + e.Reason
}

// Unwrap preserves a decoder failure for structured error inspection.
func (e *CanonicalJSONError) Unwrap() error { return e.Cause }

// CanonicalJSON serializes one JSON value according to CONFIRMATION-PROTOCOL.
// It rejects duplicate keys, floats, invalid Unicode and out-of-range integers.
// Arrays retain their input order: schema-specific set ordering, reference
// closure, subject validation and approval are responsibilities of the caller.
func CanonicalJSON(raw []byte) ([]byte, error) {
	if !utf8.Valid(raw) {
		return nil, &CanonicalJSONError{Reason: "invalid UTF-8"}
	}
	if err := checkUnicodeEscapes(raw); err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	result, err := canonicalValue(decoder)
	if err != nil {
		return nil, err
	}
	if _, err = decoder.Token(); err != io.EOF {
		return nil, &CanonicalJSONError{Reason: "trailing input", Cause: err}
	}
	return result, nil
}

// CanonicalHash returns the lowercase SHA256 of rwt-cjson-1 bytes.
// A digest establishes neither a valid subject nor an approval receipt.
func CanonicalHash(raw []byte) (string, error) {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return "", fmt.Errorf("canonical hash: %w", err)
	}
	digest := sha256.Sum256(canonical)
	return hex.EncodeToString(digest[:]), nil
}

func canonicalValue(decoder *json.Decoder) ([]byte, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, &CanonicalJSONError{Reason: "invalid JSON value", Cause: err}
	}
	switch value := token.(type) {
	case nil:
		return []byte("null"), nil
	case bool:
		return []byte(strconv.FormatBool(value)), nil
	case string:
		return appendString(nil, value), nil
	case json.Number:
		number, err := strconv.ParseInt(string(value), 10, 64)
		if err != nil {
			return nil, &CanonicalJSONError{Reason: "value is not a signed 64-bit integer", Cause: err}
		}
		return []byte(strconv.FormatInt(number, 10)), nil
	case json.Delim:
		switch value {
		case '{':
			return canonicalObject(decoder)
		case '[':
			return canonicalArray(decoder)
		}
	}
	return nil, &CanonicalJSONError{Reason: "unexpected token"}
}

func canonicalObject(decoder *json.Decoder) ([]byte, error) {
	fields := make(map[string][]byte)
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return nil, &CanonicalJSONError{Reason: "invalid object key", Cause: err}
		}
		key, ok := token.(string)
		if !ok {
			return nil, &CanonicalJSONError{Reason: "object key is not a string"}
		}
		if _, exists := fields[key]; exists {
			return nil, &CanonicalJSONError{Reason: "duplicate object key"}
		}
		value, err := canonicalValue(decoder)
		if err != nil {
			return nil, err
		}
		fields[key] = value
	}
	if err := closeContainer(decoder, '}'); err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys) // Go string ordering is lexicographic UTF-8 byte ordering.
	result := []byte{'{'}
	for i, key := range keys {
		if i > 0 {
			result = append(result, ',')
		}
		result = appendString(result, key)
		result = append(result, ':')
		result = append(result, fields[key]...)
	}
	return append(result, '}'), nil
}

func canonicalArray(decoder *json.Decoder) ([]byte, error) {
	result := []byte{'['}
	first := true
	for decoder.More() {
		value, err := canonicalValue(decoder)
		if err != nil {
			return nil, err
		}
		if !first {
			result = append(result, ',')
		}
		first = false
		result = append(result, value...)
	}
	if err := closeContainer(decoder, ']'); err != nil {
		return nil, err
	}
	return append(result, ']'), nil
}

func closeContainer(decoder *json.Decoder, expected json.Delim) error {
	token, err := decoder.Token()
	if err != nil || token != expected {
		return &CanonicalJSONError{Reason: "invalid container end", Cause: err}
	}
	return nil
}

func appendString(dst []byte, value string) []byte {
	dst = append(dst, '"')
	for _, b := range []byte(value) {
		switch b {
		case '"', '\\':
			dst = append(dst, '\\', b)
		case '\b':
			dst = append(dst, '\\', 'b')
		case '\t':
			dst = append(dst, '\\', 't')
		case '\n':
			dst = append(dst, '\\', 'n')
		case '\f':
			dst = append(dst, '\\', 'f')
		case '\r':
			dst = append(dst, '\\', 'r')
		default:
			if b < 0x20 {
				const digits = "0123456789abcdef"
				dst = append(dst, '\\', 'u', '0', '0', digits[b>>4], digits[b&15])
			} else {
				dst = append(dst, b)
			}
		}
	}
	return append(dst, '"')
}

// encoding/json replaces malformed surrogate escapes with U+FFFD. Validate
// them first, while leaving general JSON grammar validation to the decoder.
func checkUnicodeEscapes(raw []byte) error {
	inString := false
	for i := 0; i < len(raw); i++ {
		if raw[i] == '"' {
			inString = !inString
			continue
		}
		if !inString || raw[i] != '\\' {
			continue
		}
		i++
		if i >= len(raw) || raw[i] != 'u' {
			continue
		}
		unit, err := unicodeUnit(raw, i)
		if err != nil {
			return err
		}
		i += 4
		if unit >= 0xdc00 && unit <= 0xdfff {
			return &CanonicalJSONError{Reason: "unpaired low surrogate"}
		}
		if unit < 0xd800 || unit > 0xdbff {
			continue
		}
		if i+2 >= len(raw) || raw[i+1] != '\\' || raw[i+2] != 'u' {
			return &CanonicalJSONError{Reason: "unpaired high surrogate"}
		}
		low, err := unicodeUnit(raw, i+2)
		if err != nil {
			return err
		}
		if low < 0xdc00 || low > 0xdfff {
			return &CanonicalJSONError{Reason: "invalid surrogate pair"}
		}
		i += 6
	}
	return nil
}

func unicodeUnit(raw []byte, u int) (uint64, error) {
	if u+4 >= len(raw) {
		return 0, &CanonicalJSONError{Reason: "incomplete Unicode escape"}
	}
	unit, err := strconv.ParseUint(string(raw[u+1:u+5]), 16, 16)
	if err != nil {
		return 0, &CanonicalJSONError{Reason: "invalid Unicode escape", Cause: err}
	}
	return unit, nil
}
