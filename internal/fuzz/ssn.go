package fuzz

import (
	"fmt"
	"strings"
	"time"
)

func RandSSN() string {
	return SeededSSN(time.Now().UnixNano())
}

// SeededSSN generates a random valid SSN
// - First 3 digits (area number): cannot be 000, 666, or 900-999
// - Middle 2 digits (group number): cannot be 00
// - Last 4 digits (serial number): cannot be 0000
func SeededSSN(seed int64) string {
	rng := NewRandomSource(seed)

	// Generate area number (first 3 digits)
	var area int
	for {
		area = rng.Intn(900)
		if area != 0 && area != 666 && area < 900 {
			break
		}
	}

	// Generate group number (middle 2 digits)
	var group int
	for {
		group = rng.Intn(100)
		if group != 0 {
			break
		}
	}

	// Generate serial number (last 4 digits)
	var serial int
	for {
		serial = rng.Intn(10000)
		if serial != 0 {
			break
		}
	}

	return fmt.Sprintf("%03d-%02d-%04d", area, group, serial)
}

// IsValidSSN checks if an SSN is in the correct format and follows basic rules
func IsValidSSN(ssn string) bool {
	// Remove dashes if present
	ssn = strings.ReplaceAll(ssn, "-", "")

	// Check length
	if len(ssn) != 9 {
		return false
	}

	// Check if it contains only digits
	for _, char := range ssn {
		if char < '0' || char > '9' {
			return false
		}
	}

	// Parse individual parts
	area, _ := fmt.Sscanf(ssn[:3], "%d", new(int))
	group, _ := fmt.Sscanf(ssn[3:5], "%d", new(int))
	serial, _ := fmt.Sscanf(ssn[5:], "%d", new(int))

	// Validate according to SSN rules
	if area == 0 || area == 666 || (area >= 900 && area <= 999) {
		return false
	}

	if group == 0 {
		return false
	}

	if serial == 0 {
		return false
	}

	return true
}
