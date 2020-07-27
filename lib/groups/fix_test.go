package groups

import (
	"testing"
)

func Test_cleanGroup(t *testing.T) {
	tests := []struct {
		name   string
		wantOk bool
	}{
		{"", false},
		{"Defacto2", false},
		{"defacto2", true},
		{"d-e-f-a-c-t-o-2", true},
		{"d_f", true},
		{"D2", false},
	}
	sim = false
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotOk := cleanGroup(tt.name); gotOk != tt.wantOk {
				t.Errorf("cleanGroup() = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func Test_cleanString(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"", args{""}, ""},
		{"", args{"the blah"}, "The Blah"},
		{"", args{"in the blah"}, "In the Blah"},
		{"", args{"TheBlah"}, "Theblah"},
		{"", args{"MiRROR now"}, "Mirror Now"},
		{"", args{"In the row now ii"}, "In the Row Now II"},
		{"", args{"MiRROR now bbS"}, "Mirror Now BBS"},
		{"", args{"this-is-a-slug-string"}, "This-Is-A-Slug-String"},
		{"", args{"Group inc.,RAZOR TO 1911"}, "Group Inc,Razor to 1911"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanString(tt.args.s); got != tt.want {
				t.Errorf("cleanString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_trimSP(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"", args{"abc"}, "abc"},
		{"", args{"a b c"}, "a b c"},
		{"", args{"a  b  c"}, "a b c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := trimSP(tt.args.s); got != tt.want {
				t.Errorf("trimSP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_trimThe(t *testing.T) {
	type args struct {
		g string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"", args{"The X BBS"}, "X BBS"},
		{"", args{"The X FTP"}, "X FTP"},
		{"", args{"the X BBS"}, "X BBS"},
		{"", args{"THE X BBS"}, "X BBS"},
		{"", args{"The"}, "The"},
		{"", args{"Hello BBS"}, "Hello BBS"},
		{"", args{"The High & Mighty Hello BBS"}, "High & Mighty Hello BBS"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := trimThe(tt.args.g); got != tt.want {
				t.Errorf("trimThe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_trimDot(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{""}, ""},
		{"no dots", args{"hello"}, "hello"},
		{"dots", args{"hello."}, "hello"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := trimDot(tt.args.s); got != tt.want {
				t.Errorf("trimDot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_format(t *testing.T) {
	tests := []struct {
		s    string
		want string
	}{
		{"", ""},
		{"hello", "Hello"},
		{"hello  world", "Hello  World"},
		{"By THE Way", "By the Way"},
		{"BENS ftp", "Bens FTP"},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			if got := format(tt.s); got != tt.want {
				t.Errorf("format() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFix(t *testing.T) {
	tests := []struct {
		name     string
		simulate bool
		wantErr  bool
	}{
		{"sim", true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Fix(tt.simulate); (err != nil) != tt.wantErr {
				t.Errorf("Fix() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
