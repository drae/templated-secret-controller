// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package generator

import "testing"

func TestNamespaceAuthorized(t *testing.T) {
	cases := []struct {
		value  string
		ns     string
		expect bool
	}{
		{"", "a", false},
		{"team-a", "team-a", true},
		{"team-a,team-b", "team-c", false},
		{" team-a , team-b ", "team-b", true},
		{"*", "x", true},
		{"team-a,*", "other", true},
	}
	for _, c := range cases {
		if got := namespaceAuthorized(c.value, c.ns); got != c.expect {
			// include case value for debugging
			t.Fatalf("namespaceAuthorized(%q,%q) expected %v got %v", c.value, c.ns, c.expect, got)
		}
	}
}
