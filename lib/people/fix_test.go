package people

import "testing"

func Test_cleanPerson(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"", args{""}, ""},
		{"", args{"the blah"}, "the blah"},
		{"", args{"a dude,blah"}, "a dude,blah"},
		{"", args{"name1,name2,!name3!"}, "name1,name2,name3!"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanPerson(tt.args.s); got != tt.want {
				t.Errorf("cleanPerson() = %v, want %v", got, tt.want)
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
		{"", args{"a nick"}, "a nick"},
		{"", args{"--a nick"}, "a nick"},
		{"", args{" ?!nick!! "}, "nick"},
		{"", args{"?!nick!!,someone else"}, "nick,someone else"},
		{"", args{"?!nick!!,--someone-else++"}, "nick,--someone-else"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanString(tt.args.s); got != tt.want {
				t.Errorf("cleanString() = %v, want %v", got, tt.want)
			}
		})
	}
}
