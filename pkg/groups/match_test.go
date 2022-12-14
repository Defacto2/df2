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

func TestSwapOne(t *testing.T) {
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
		{"multi", args{"the three styling", "th", "f"}, "Fe Three Styling"},
		{"emoji", args{"do emojis work?", "work", "👉🏿"}, "Do Emojis 👉🏿?"},
		{"case", args{"heLLo", "l", "1"}, "He1lo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SwapOne(tt.args.group, tt.args.phonetic, tt.args.swap); got != tt.want {
				t.Errorf("SwapOne() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSwapAll(t *testing.T) {
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
		{"zeros", args{"hell0 H00T", "0", "o"}, "Hello Hoot"},
		{"case", args{"heLLo", "l", "1"}, "He11o"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SwapAll(tt.args.group, tt.args.phonetic, tt.args.swap); got != tt.want {
				t.Errorf("SwapAll() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSwapNumeral(t *testing.T) {
	type args struct {
		group string
		i     int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{}, ""},
		{"one", args{"razor 1", 1}, "Razor One"},
		{"zero", args{"0razor", 0}, "Zerorazor"},
		{"twelve", args{"i am 12", 12}, "I AM Twelve"},
		{"100 out of range", args{"100 pounds", 100}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SwapNumeral(tt.args.group, tt.args.i); got != tt.want {
				t.Errorf("SwapNumeral() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSwapOrdinal(t *testing.T) {
	type args struct {
		group string
		i     int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{}, ""},
		{"one", args{"razor 1", 1}, "Razor 1st"},
		{"zero", args{"0razor", 0}, "0Razor"},
		{"twelve", args{"i am 12", 12}, "I AM 12th"},
		{"100 out of range", args{"100 pounds", 100}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SwapOrdinal(tt.args.group, tt.args.i); got != tt.want {
				t.Errorf("SwapOrdinal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSwapPrefix(t *testing.T) {
	type args struct {
		group  string
		prefix string
		swap   string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{}, ""},
		{"lcase", args{"the best group", "the", "da"}, "Da Best Group"},
		{"mix case", args{"The best BBS", "the", "da"}, "Da Best BBS"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SwapPrefix(tt.args.group, tt.args.prefix, tt.args.swap); got != tt.want {
				t.Errorf("SwapPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSwapSuffix(t *testing.T) {
	type args struct {
		group  string
		suffix string
		swap   string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{}, ""},
		{"lcase", args{"the best group", "group", "crew"}, "The Best Crew"},
		{"mix case", args{"OLDSKOOLS", "s", "z"}, "Oldskoolz"},
		{"ftp", args{"The 1ST FTP", "st", "XX"}, "The 1Xx FTP"},
		{"unicode", args{"apple=🍏", "🍏", "🍎"}, "Apple=🍎"},
		{"unicode+bbs", args{"🍏 bbs", "🍏", "🍎"}, "🍎 BBS"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SwapSuffix(tt.args.group, tt.args.suffix, tt.args.swap); got != tt.want {
				t.Errorf("SwapSuffix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrimSP(t *testing.T) {
	tests := []struct {
		name  string
		s     string
		want  string
		want1 string
		want2 string
	}{
		{"empty", "", "", "s", "z"},
		{"str", "bon bon", "Bonbon", "Bonbons", "Bonbonz"},
		{"bbs", "BON bon BBS", "Bonbon BBS", "Bonbons BBS", "Bonbonz BBS"},
		{"ftp", "A BON bon FTP", "Abonbon FTP", "Abonbons FTP", "Abonbonz FTP"},
		{"unicode", "🍎 apples 🍏", "🍎Apples🍏", "🍎Apples🍏s", "🍎Apples🍏z"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := TrimSP(tt.s)
			if got != tt.want {
				t.Errorf("TrimSP() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("TrimSP() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("TrimSP() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}
