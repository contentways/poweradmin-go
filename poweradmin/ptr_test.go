// Copyright (c) 2026 Contentways
// SPDX-License-Identifier: MIT
package poweradmin

import "testing"

func TestPtr(t *testing.T) {
	s := Ptr("x")
	if s == nil || *s != "x" {
		t.Errorf("Ptr(\"x\") = %v", s)
	}
	n := Ptr(0)
	if n == nil || *n != 0 {
		t.Errorf("Ptr(0) = %v", n)
	}
}
