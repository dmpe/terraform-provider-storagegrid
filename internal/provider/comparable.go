// Copyright github.com/dmpe 2024, 2026
// SPDX-License-Identifier: MIT

package provider

import (
	"strings"
	"time"
)

// EqualElements checks if two slices contain the same elements, regardless of order.
func EqualElements[T comparable](s1, s2 []T) bool {
	if len(s1) != len(s2) {
		return false
	}

	counts := make(map[T]int)
	for _, v := range s1 {
		counts[v]++
	}

	for _, v := range s2 {
		if counts[v] == 0 {
			// Element kommt in s2 öfter vor oder existiert in s1 gar nicht
			return false
		}
		counts[v]--
	}

	return true
}

// EqualDates compares two date strings in RFC3339 format with or without milliseconds.
// Returns true if year, month, day, hour, minute and second are equal.
func EqualDates(date1, date2 string) bool {
	// Normalize dates by removing milliseconds if present
	normalizeDate := func(date string) string {
		if strings.Contains(date, ".") {
			// Remove milliseconds: "2026-07-16T14:17:07.000Z" -> "2026-07-16T14:17:07Z"
			parts := strings.Split(date, ".")
			if len(parts) == 2 {
				return parts[0] + "Z"
			}
		}
		return date
	}

	normalized1 := normalizeDate(date1)
	normalized2 := normalizeDate(date2)

	t1, err1 := time.Parse(time.RFC3339, normalized1)
	t2, err2 := time.Parse(time.RFC3339, normalized2)

	if err1 != nil || err2 != nil {
		return false
	}

	return t1.Equal(t2)
}
