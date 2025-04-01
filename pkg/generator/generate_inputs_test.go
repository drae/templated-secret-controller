// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package generator_test

import (
	"errors"
	"testing"

	"github.com/drae/templated-secret-controller/pkg/generator"
	"github.com/stretchr/testify/assert"
)

func TestAddFailsWithEmptyAnnotations(t *testing.T) {
	err := generator.GenerateInputs{}.Add(nil)
	assert.Equal(t, errors.New("internal inconsistency: called with annotations nil param"), err)
}

func TestAddSucceedsfulWithDefaultAnnotation(t *testing.T) {
	defaultAnnotations := map[string]string{
		"templatedsecret.starstreak.dev/generate-inputs": "",
	}
	err := generator.GenerateInputs{}.Add(defaultAnnotations)
	assert.Equal(t, nil, err)
}
