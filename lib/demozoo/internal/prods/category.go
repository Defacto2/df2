package prods

import "strings"

// Category are tags for production imports.
type Category int

func (c Category) String() string {
	switch c {
	case Text:
		return "text"
	case Code:
		return "code"
	case Graphics:
		return "graphics"
	case Music:
		return "music"
	case Magazine:
		return "magazine"
	}
	return ""
}

const (
	// Text based files.
	Text Category = iota
	// Code are binary files.
	Code
	// Graphics are images.
	Graphics
	// Music is audio.
	Music
	// Magazine are publications.
	Magazine
)

func category(s string) Category {
	switch strings.ToLower(s) {
	case Text.String():
		return Text
	case Code.String():
		return Code
	case Graphics.String():
		return Graphics
	case Music.String():
		return Music
	case Magazine.String():
		return Magazine
	}
	return -1
}
