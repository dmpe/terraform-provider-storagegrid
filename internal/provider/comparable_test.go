// Copyright github.com/dmpe 2024, 2026
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

func Test_EqualDates(t *testing.T) {
	tests := []struct {
		name    string
		aDate   string
		bDate   string
		isEqual bool
	}{
		{
			name:    "EqualDatesWithMilliseconds",
			aDate:   "2026-07-16T14:17:07.000Z",
			bDate:   "2026-07-16T14:17:07.000Z",
			isEqual: true,
		},
		{
			name:    "DifferentDatesWithMilliseconds",
			aDate:   "2026-07-16T14:17:07.000Z",
			bDate:   "2026-07-15T14:17:07.000Z",
			isEqual: false,
		},
		{
			name:    "DifferentTimesOnSameDate",
			aDate:   "2026-07-16T14:17:07.000Z",
			bDate:   "2026-07-16T15:45:30.000Z",
			isEqual: false,
		},
		{
			name:    "DifferentDatesWithoutMilliseconds",
			aDate:   "2026-07-16T10:30:00Z",
			bDate:   "2026-07-15T10:30:00Z",
			isEqual: false,
		},
		{
			name:    "EqualDatesWithDifferentFormats",
			aDate:   "2026-07-16T10:30:00Z",
			bDate:   "2026-07-16T10:30:00Z",
			isEqual: true,
		},
		{
			name:    "EqualDatesWithDifferentFormatsAndMilliseconds",
			aDate:   "2026-07-16T10:30:00.000Z",
			bDate:   "2026-07-16T10:30:00Z",
			isEqual: true,
		},
		{
			name:    "EqualDateTimeWithDifferentMilliseconds",
			aDate:   "2026-07-16T14:17:07.000Z",
			bDate:   "2026-07-16T14:17:07.921Z",
			isEqual: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.isEqual, EqualDates(test.aDate, test.bDate))
		})
	}
}
