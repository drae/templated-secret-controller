//go:build integration
// +build integration

// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package ci

import (
	"os"
	"strings"
	"testing"
)

type Env struct {
	Namespace string
}

func BuildEnv(t *testing.T) Env {
	env := Env{
		Namespace: os.Getenv("NAMESPACE"),
	}
	env.Validate(t)
	return env
}

func (e Env) Validate(t *testing.T) {
	errStrs := []string{}

	if len(e.Namespace) == 0 {
		errStrs = append(errStrs, "Expected Namespace to be non-empty")
	}

	if len(errStrs) > 0 {
		t.Fatalf("%s", strings.Join(errStrs, "\n"))
	}
}
