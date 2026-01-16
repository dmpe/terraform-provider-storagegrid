// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MIT

package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_EqualElements(t *testing.T) {
	a := []string{"Apfel", "Birne", "Banane"} // baseline
	b := []string{"Banane", "Apfel", "Birne"} // same value, differently sorted
	c := []string{"Apfel", "Apfel", "Banane"} // different values
	d := []string{"Apfel", "Birne"}           // different length

	assert.True(t, EqualElements(a, b))
	assert.False(t, EqualElements(a, c))
	assert.False(t, EqualElements(a, d))
}
