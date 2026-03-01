// Copyright (c) 2026 Kartoza
// SPDX-License-Identifier: MIT

//go:build ignore
// +build ignore

// This is a helper program to generate icon files
// Run with: go run generate_icons.go

package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

func main() {
	// Create green icon
	createIcon("green.png", color.RGBA{0, 200, 0, 255})

	// Create orange icon
	createIcon("orange.png", color.RGBA{255, 165, 0, 255})

	// Create red icon
	createIcon("red.png", color.RGBA{255, 0, 0, 255})
}

func createIcon(filename string, col color.Color) {
	const size = 64
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Fill with transparent background
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, color.RGBA{0, 0, 0, 0})
		}
	}

	// Draw a filled circle
	center := size / 2
	radius := size/2 - 4

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := x - center
			dy := y - center
			if dx*dx+dy*dy <= radius*radius {
				img.Set(x, y, col)
			}
		}
	}

	// Save to file
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
}
