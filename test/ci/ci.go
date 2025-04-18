//go:build integration
// +build integration

// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package ci

import (
	"fmt"
)

type Logger struct{}

func (l Logger) Section(msg string, f func()) {
	fmt.Printf("==> %s\n", msg)
	f()
}

func (l Logger) Debugf(msg string, args ...interface{}) {
	fmt.Printf(msg, args...)
}
