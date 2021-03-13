package database

import (
	"testing"
)

func Test_StripChars(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"", args{""}, ""},
		{"", args{"ooÖØöøO"}, "ooÖØöøO"},
		{"", args{"o.o|Ö+Ø=ö^ø#O"}, "ooÖØöøO"},
		{"", args{"A Café!"}, "A Café"},
		{"", args{"brunräven - över"}, "brunräven - över"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StripChars(tt.args.s); got != tt.want {
				t.Errorf("StripChars() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_StripStart(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"", args{""}, ""},
		{"", args{"hello world"}, "hello world"},
		{"", args{"--argument"}, "argument"},
		{"", args{"!!!OMG-WTF"}, "OMG-WTF"},
		{"", args{"#ÖØöøO"}, "ÖØöøO"},
		{"", args{"!@#$%^&A(+)ooÖØöøO"}, "A(+)ooÖØöøO"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StripStart(tt.args.s); got != tt.want {
				t.Errorf("StripStart() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_TrimSP(t *testing.T) {
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
			if got := TrimSP(tt.args.s); got != tt.want {
				t.Errorf("TrimSP() = %v, want %v", got, tt.want)
			}
		})
	}
}
