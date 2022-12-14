package groups

import (
	"sort"
	"testing"
)

func TestContains(t *testing.T) {
	const hi = "👋 hi!"
	x := []string{"hello", "world", "apple", "banana", "carrot", hi, "cake"}
	sort.Strings(x)
	type args struct {
		x string
		s []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"empty", args{}, false},
		{"blank", args{"", x}, false},
		{"no match", args{"abcde", x}, false},
		{"match", args{"apple", x}, true},
		{"unicode", args{hi, x}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Contains(tt.args.x, tt.args.s); got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSwapPhonetic(t *testing.T) {
	type args struct {
		group    string
		phonetic string
		swap     string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{}, ""},
		{"none", args{"hello", "th", "f"}, "Hello"},
		{"prefix", args{"threestyling", "th", "f"}, "Freestyling"},
		{"sentence", args{"A threestyler", "th", "f"}, "A Freestyler"},
		{"multi", args{"the three styling", "th", "f"}, "Fe Free Styling"},
		{"emoji", args{"do emojis work?", "work", "👉🏿"}, "Do Emojis 👉🏿?"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SwapPhonetic(tt.args.group, tt.args.phonetic, tt.args.swap); got != tt.want {
				t.Errorf("SwapPhonetic() = %v, want %v", got, tt.want)
			}
		})
	}
}
