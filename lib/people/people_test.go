package people

import (
	"reflect"
	"strings"
	"testing"
)

func Test_roles(t *testing.T) {
	type args struct {
		r string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"", args{""}, "wmca"},
		{"", args{"artists"}, "a"},
		{"", args{"a"}, "a"},
		{"", args{"all"}, "wmca"},
		{"error", args{"xxx"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := roles(tt.args.r); got != tt.want {
				t.Errorf("roles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_peopleStmt(t *testing.T) {
	type args struct {
		role               string
		includeSoftDeletes bool
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"error", args{"text", false}, 0},
		{"empty f", args{"", false}, 507},
		{"empty t", args{"", true}, 415},
		{"writers", args{"writers", true}, 109},
		{"writers", args{"w", true}, 109},
		{"musicians", args{"m", true}, 111},
		{"coders", args{"c", true}, 115},
		{"artists", args{"a", true}, 125},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := peopleStmt(tt.args.role, tt.args.includeSoftDeletes); len(got) != tt.want {
				t.Errorf("sqlPeople() = %v, want = %v", len(got), tt.want)
			}
		})
	}
}

func TestWheres(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{"", strings.Split(Filters, ",")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Wheres(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Wheres() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestList(t *testing.T) {
	tests := []struct {
		name string
		role string
		want int
	}{
		{"empty", "", 1},
		{"error", "error", 0},
		{"writers", "writers", 1},
		{"musicians", "m", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got, err := List(tt.role)
			if err != nil {
				t.Error(err)
			}
			if got < tt.want {
				t.Errorf("List() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrint(t *testing.T) {
	tests := []struct {
		name string
		r    Request
	}{
		{"empty", Request{}},
		{"regular", Request{"writer", false, true}},
		{"error", Request{"error", false, true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Print(tt.r)
		})
	}
}

func TestDataList(t *testing.T) {
	type args struct {
		filename string
		r        Request
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"error", args{"", Request{"error", false, false}}, true},
		{"ok", args{"", Request{"", false, false}}, false},
		{"progress", args{"", Request{"", false, true}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DataList(tt.args.filename, tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("DataList() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHTML(t *testing.T) {
	type args struct {
		filename string
		r        Request
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"error", args{"", Request{"error", false, false}}, true},
		{"ok", args{"", Request{"", false, false}}, false},
		{"progress", args{"", Request{"", false, true}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := HTML(tt.args.filename, tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("HTML() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
