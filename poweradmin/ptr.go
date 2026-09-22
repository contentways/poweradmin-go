// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

// Ptr returns a pointer to v. It is a convenience for the optional pointer
// fields in the *UpdateOpts types:
//
//	client.Record.Update(ctx, zoneID, recordID, poweradmin.RecordUpdateOpts{
//		Content: poweradmin.Ptr("192.0.2.10"),
//		TTL:     poweradmin.Ptr(300),
//	})
func Ptr[T any](v T) *T {
	return &v
}
