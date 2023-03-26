package groups_test

import (
	"io"
	"sort"
	"testing"

	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/groups"
	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	t.Parallel()
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
			if got := groups.Contains(tt.args.x, tt.args.s); got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatch(t *testing.T) {
	t.Parallel()
	err := groups.Match(nil, nil, -1)
	assert.NotNil(t, err)

	db, err := database.Connect(conf.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = groups.Match(db, io.Discard, 100)
	assert.Nil(t, err)
}

func TestSwapOne(t *testing.T) {
	t.Parallel()
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
			if got := groups.SwapOne(tt.args.group, tt.args.phonetic, tt.args.swap); got != tt.want {
				t.Errorf("SwapOne() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSwapAll(t *testing.T) {
	t.Parallel()
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
			if got := groups.SwapAll(tt.args.group, tt.args.phonetic, tt.args.swap); got != tt.want {
				t.Errorf("SwapAll() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSwapNumeral(t *testing.T) {
	t.Parallel()
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
			if got := groups.SwapNumeral(tt.args.group, tt.args.i); got != tt.want {
				t.Errorf("SwapNumeral() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSwapOrdinal(t *testing.T) {
	t.Parallel()
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
			if got := groups.SwapOrdinal(tt.args.group, tt.args.i); got != tt.want {
				t.Errorf("SwapOrdinal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSwapPrefix(t *testing.T) {
	t.Parallel()
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
			if got := groups.SwapPrefix(tt.args.group, tt.args.prefix, tt.args.swap); got != tt.want {
				t.Errorf("SwapPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSwapSuffix(t *testing.T) {
	t.Parallel()
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
			if got := groups.SwapSuffix(tt.args.group, tt.args.suffix, tt.args.swap); got != tt.want {
				t.Errorf("SwapSuffix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrimSP(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		s     string
		want  string
		want1 string
		want2 string
	}{
		{"empty", "", "", "s", "z"},
		{"str", "bon bon", "Bonbon", "Bonbons", "Bonbonz"}, //nolint:dupword
		{"bbs", "BON bon BBS", "Bonbon BBS", "Bonbons BBS", "Bonbonz BBS"},
		{"ftp", "A BON bon FTP", "Abonbon FTP", "Abonbons FTP", "Abonbonz FTP"},
		{"unicode", "🍎 apples 🍏", "🍎Apples🍏", "🍎Apples🍏s", "🍎Apples🍏z"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := groups.TrimSP(tt.s)
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
