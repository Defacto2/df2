package str_test

import (
	"testing"

	"github.com/Defacto2/df2/pkg/str"
	"github.com/gookit/color"
	"github.com/stretchr/testify/assert"
)

func capString(test string) string {
	color.Enable = false
	switch test {
	case "x":
		return str.X()
	case "y":
		return str.Y()
	}
	return ""
}

func Test_capString(t *testing.T) {
	t.Parallel()
	type args struct {
		test string
	}
	tests := []struct {
		name       string
		args       args
		wantOutput string
	}{
		{"empty", args{"sec"}, ""},
		{"x", args{"x"}, "✗"},
		{"y", args{"y"}, "✓"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if gotOutput := capString(tt.args.test); gotOutput != tt.wantOutput {
				t.Errorf("capString() = %v, want %v", gotOutput, tt.wantOutput)
			}
		})
	}
}

func TestProgress(t *testing.T) {
	t.Parallel()
	type args struct {
		name  string
		count int
		total int
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{"ten", args{"", 1, 10}, float64(10)},
		{"hundred", args{"", 10, 10}, float64(100)},
		{"zero", args{"", 0, 10}, float64(0)},
		{"negative", args{"", -1, 10}, float64(-10)},
		{"decimal", args{"", 1, 99999}, float64(0.001000010000100001)},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := str.Progress(nil, tt.args.name, tt.args.count, tt.args.total); got != tt.want {
				t.Errorf("Progress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	t.Parallel()
	type args struct {
		text string
		len  int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{"", 0}, ""},
		{"zero", args{"hello", 0}, "hello"},
		{"minus", args{"hello", -10}, "hello"},
		{"three", args{"hello", 3}, "he…"},
		{"too long", args{"hello", 600}, "hello"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := str.Truncate(tt.args.text, tt.args.len); got != tt.want {
				t.Errorf("Truncate() = %v, want %v", got, tt.want)
			}
		})
	}
}

const (
	r1 = "Acronis.Disk.Director.Suite.v10.0.2077.Russian.Incl.Keymaker-ZWT"
	r2 = "Apollo-tech.No1.Video.Converter.v3.8.17.Incl.Keymaker-ZWT"
	r3 = "SiSoftware.Sandra.Pro.Business.XI.SP3.2007.6.11.40.Multilingual.Retail.Incl.Keymaker-ZWT"
)

func TestTitle(t *testing.T) {
	t.Parallel()
	s := str.PathTitle("")
	assert.Equal(t, "", s)
	s = str.PathTitle("HeLLo worLD! ")
	assert.Equal(t, "HeLLo worLD!", s)
	s = str.PathTitle(r1)
	assert.Equal(t, "Acronis Disk Director Suite v10.0.2077 Russian including keymaker", s)
	s = str.PathTitle("Acronis.Disk.Director.Suite.v10.1.Russian.Incl.Keymaker-ZWT")
	assert.Equal(t, "Acronis Disk Director Suite v10.1 Russian including keymaker", s)
	s = str.PathTitle("Acronis.Disk.Director.Suite.v10.Russian.Incl.Keymaker-ZWT")
	assert.Equal(t, "Acronis Disk Director Suite v10 Russian including keymaker", s)
	s = str.PathTitle(r2)
	assert.Equal(t, "Apollo-tech No1 Video Converter v3.8.17 including keymaker", s)
	s = str.PathTitle(r3)
	assert.Equal(t, "SiSoftware Sandra Pro Business XI SP3 2007 6 11 40 Multilingual Retail including keymaker", s)
}
