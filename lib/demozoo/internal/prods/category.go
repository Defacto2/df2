package prods

import "strings"

// Category are tags for production imports.
type Category int

const (
	Text     Category = iota // Text based files.
	Code                     // Code are binary files.
	Graphics                 // Graphics are images.
	Music                    // Music is audio.
	Magazine                 // Magazine are publications.
)

func (c Category) String() string {
	return [...]string{"text", "code", "graphics", "music", "magazine"}[c]
}

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
