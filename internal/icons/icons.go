// Copyright (c) 2026 Kartoza
// SPDX-License-Identifier: MIT

// Package icons provides embedded icon resources for different status states
package icons

import (
	_ "embed"
)

// GreenIcon is the icon for healthy status (all endpoints OK)
//
//go:embed green.png
var GreenIcon []byte

// OrangeIcon is the icon for warning status (1-2 errors)
//
//go:embed orange.png
var OrangeIcon []byte

// RedIcon is the icon for error status (3+ errors or unreachable)
//
//go:embed red.png
var RedIcon []byte

// GetIconForStatus returns the appropriate icon for the given status string
func GetIconForStatus(status string) []byte {
	switch status {
	case "green":
		return GreenIcon
	case "orange":
		return OrangeIcon
	case "red":
		return RedIcon
	default:
		return GreenIcon
	}
}
