// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MIT

package provider

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
