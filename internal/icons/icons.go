// Package icons provides embedded icon resources for different status states
package icons

import (
	_ "embed"
)

// Icon byte slices for different status states
// These are simple PNG icons that work across all platforms

//go:embed green.png
var GreenIcon []byte

//go:embed orange.png
var OrangeIcon []byte

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
