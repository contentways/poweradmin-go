// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT

package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// RecordID is the wire representation of a record identifier.
//
// Poweradmin normalizes record IDs before returning them: purely numeric IDs
// (the default for MySQL, PostgreSQL and SQLite backends) are emitted as JSON
// numbers, everything else as JSON strings. The OpenAPI spec documents this as
// oneOf integer|string. RecordID accepts both forms and always holds the ID as
// its decimal/string representation, so the domain layer can keep using a
// plain string.
type RecordID string

// UnmarshalJSON accepts a JSON number, a JSON string or null.
func (id *RecordID) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if bytes.Equal(data, []byte("null")) {
		*id = ""
		return nil
	}
	if len(data) > 0 && data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return fmt.Errorf("schema: record id: %w", err)
		}
		*id = RecordID(s)
		return nil
	}
	var n json.Number
	if err := json.Unmarshal(data, &n); err != nil {
		return fmt.Errorf("schema: record id must be a number or string, got %s", data)
	}
	*id = RecordID(n.String())
	return nil
}
